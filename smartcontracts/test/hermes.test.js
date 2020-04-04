// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

const Hermes = artifacts.require('Hermes.sol');
const MultiSend = artifacts.require('MultiSendMock.sol')
const ForwardRegistration = artifacts.require('ForwardRegistration.sol');

const MinTips = 9876543210;
const Limit = 2;
const ContractStartEpoch = 1000;
const AnalyticsEndpoint = "https://analytics.iotexscan.io/";

// Test Accounts
// Accounts:
//     (0) 0x627306090abab3a6e1400e9345bc60c78a8bef57
//     (1) 0xf17f52151ebef6c7334fad080c5704d77216b732
//     (2) 0xc5fdf4076b8f3a5357c5e395ab970b5b54098fef
//     (3) 0x821aea9a577a9b44299b9c15c88cf3087f3b5544
//     (4) 0x0d1d4e623d10f9fba5db95830f7d3839406c6af2
//     (5) 0x2932b7a2355d6fecc4b5c0b6bd44cc31df247a2e
//     (6) 0x2191ef87e392377ec08e7c08eb105ef5448eced5
//     (7) 0x0f4f2ac550a1b4e2280d04c21cea7ebd822934b5
//     (8) 0x6330a553fc93768f612722bb8c2ec78ac90b3bbc
//     (9) 0x5aeda56215b167893e80b4fe645ba6d5bab767de
//
// Private Keys:
//     (0) c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3
//     (1) ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f
//     (2) 0dbbe8e4ae425a6d2687f1a7e3ba17bc98c673636790f1b8ad91193c05875ef1
//     (3) c88b703fb08cbea894b6aeff5a544fb92e78a18e19814cd85da83b71f772aa6c
//     (4) 388c684f0ba1ef5017716adb5d21a053ea8e90277d0868337519f97bede61418
//     (5) 659cbb0e2411a44db63778987b1e22153c086a95eb6b18bdf89de078917abc63
//     (6) 82d052c865f5763aad42add438569276c00d3d88a2d062d36b2bae914d58b8c8
//     (7) aa3680d5d48a8283413f7a108367c7299ca73f553735860a87b08f39395618b7
//     (8) 0f62d96d6675f32685bbdb8ac13cda7c23436f63efbb9d07700d8669ff12b7c4
//     (9) 8d5366123cb560bb606379f90a0bfd4769eecc0557f1b362dcae9012b548b1e5

contract('hermes', function(accounts) {
    beforeEach(async function() {
        this.multisend = await MultiSend.new(MinTips, Limit);
        this.forwardRegistration = await ForwardRegistration.new();
        this.contract = await Hermes.new(ContractStartEpoch, this.multisend.address, this.forwardRegistration.address, AnalyticsEndpoint);
    });
    describe('get analytics endpoint', function () {
        it('success', async function() {
            let endpoint = await this.contract.analyticsEndpoint();
            assert.equal(endpoint, AnalyticsEndpoint);
        });
    });
    describe('set analytics endpoint', function () {
        it('success', async function() {
            await this.contract.setAnalyticsEndpoint("http://iotex-analytics.herokuapp.com");
            let endpoint = await this.contract.analyticsEndpoint();
            assert.equal(endpoint, "http://iotex-analytics.herokuapp.com");
        });
    });
    describe('distribute rewards', function () {
        it('distribute to multiple addresses', async function() {
            let balances = [];
            for (let i = 0; i < 7; i++) {
                let balance = await web3.eth.getBalance(accounts[i]);
                balances.push(web3.utils.toBN(balance));
            }
            let delegate1 = web3.utils.fromAscii("iosg")
            let delegate2 = web3.utils.fromAscii("cpc")
            // first delegate
            // first group
            await this.contract.distributeRewards(delegate1, 1600, [accounts[2], accounts[3]], [100000000, 200000000], {from: accounts[0], value: MinTips + 300000000});
            assert.equal(await web3.eth.getBalance(accounts[2]), balances[2].add(web3.utils.toBN(100000000)).toString());
            assert.equal(await web3.eth.getBalance(accounts[3]), balances[3].add(web3.utils.toBN(200000000)).toString());
            let endEpochsCount = await this.contract.getEndEpochCount();
            assert.equal(endEpochsCount, 0);
            let recipientEpoch = await this.contract.recipientEpochTracker(delegate1, accounts[2]);
            let distributedCount = await this.contract.distributedCount(delegate1);
            let distributedAmount = await this.contract.distributedAmount(delegate1);
            assert.equal(recipientEpoch, 1600);
            assert.equal(distributedCount, 2);
            assert.equal(distributedAmount, 300000000);

            let distribution = await this.contract.distributions(delegate1, 1600);
            assert.equal(distribution.distributedCount, 0)
            assert.equal(distribution.amount, 0)

            // duplicate group
            let err;
            try {
                await this.contract.distributeRewards(delegate1, 1600, [accounts[2], accounts[3]], [100000000, 200000000], {from: accounts[0], value: MinTips + 300000000});
            } catch (e) {
                err = e;
            }
            assert.ok(err.toString().includes("reward has already been distributed to the recipient"))

            // second group
            await this.contract.distributeRewards(delegate1, 1600, [accounts[1], accounts[4]], [300000000, 400000000], {from: accounts[0], value: MinTips + 700000000});
            assert.equal(await web3.eth.getBalance(accounts[1]), balances[1].add(web3.utils.toBN(300000000)).toString());
            assert.equal(await web3.eth.getBalance(accounts[4]), balances[4].add(web3.utils.toBN(400000000)).toString());
            endEpochsCount = await this.contract.getEndEpochCount();
            assert.equal(endEpochsCount, 0);
            recipientEpoch = await this.contract.recipientEpochTracker(delegate1, accounts[1]);
            distributedCount = await this.contract.distributedCount(delegate1);
            distributedAmount = await this.contract.distributedAmount(delegate1);
            assert.equal(recipientEpoch, 1600);
            assert.equal(distributedCount, 4);
            assert.equal(distributedAmount, 1000000000);

            distribution = await this.contract.distributions(delegate1, 1600);
            assert.equal(distribution.distributedCount, 0);
            assert.equal(distribution.amount, 0);

            // second delegate
            await this.contract.distributeRewards(delegate2, 1600, [accounts[5], accounts[6]], [500000000, 600000000], {from: accounts[0], value: MinTips + 1100000000});
            assert.equal(await web3.eth.getBalance(accounts[5]), balances[5].add(web3.utils.toBN(500000000)).toString());
            assert.equal(await web3.eth.getBalance(accounts[6]), balances[6].add(web3.utils.toBN(600000000)).toString());
            recipientEpoch = await this.contract.recipientEpochTracker(delegate2, accounts[5]);
            distributedCount = await this.contract.distributedCount(delegate2);
            distributedAmount = await this.contract.distributedAmount(delegate2);
            assert.equal(recipientEpoch, 1600);
            assert.equal(distributedCount, 2);
            assert.equal(distributedAmount, 1100000000);

            // commit distributions
            await this.contract.commitDistributions(1600, [delegate1, delegate2], {from: accounts[0]});

            endEpochsCount = await this.contract.getEndEpochCount();
            assert.equal(endEpochsCount, 1);
            let lastEpoch = await this.contract.endEpochs(endEpochsCount-1);
            recipientEpoch = await this.contract.recipientEpochTracker(delegate2, accounts[5]);
            distributedCount = await this.contract.distributedCount(delegate2);
            distributedAmount = await this.contract.distributedAmount(delegate2);
            assert.equal(lastEpoch, 1600);
            assert.equal(recipientEpoch, 1600);
            assert.equal(distributedCount, 0);
            assert.equal(distributedAmount, 0);

            distribution = await this.contract.distributions(delegate2, 1600);
            assert.equal(distribution.distributedCount.toNumber(), 2);
            assert.equal(distribution.amount.toNumber(), 1100000000);
        });
    });
});
