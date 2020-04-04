// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

const ForwardRegistration = artifacts.require('ForwardRegistration.sol');
const MultiSend = artifacts.require('MultiSendMock.sol');
const Hermes = artifacts.require('Hermes.sol');

const MinTips = 9876543210;
const Limit = 2;
const ContractStartEpoch = 1000;
const AnalyticsEndpoint = "https://analytics.iotexscan.io/";

contract('forwardRegistration', function(accounts) {
    beforeEach(async function() {
        this.multisend = await MultiSend.new(MinTips, Limit);
        this.contract = await ForwardRegistration.new();
        this.hermes = await Hermes.new(ContractStartEpoch, this.multisend.address, this.contract.address, AnalyticsEndpoint);
    });
    describe('forward service', function () {
        it('claim as', async function () {
            // register forward service
            const registerNonce = 2;
            let registerMsgHash = registerNonce + 'I authorize ' + accounts[2].toLowerCase() + ' to register in ' + this.contract.address.toLowerCase();
            let sig = await web3.eth.accounts.sign(registerMsgHash, '0xae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f');

            sig = sig.signature;

            // succeed to register
            await this.contract.registerForwardService(accounts[1], registerNonce, 1500, sig, {from: accounts[2]});
            let balances = [];
            for (let i = 0; i < 7; i++) {
                let balance = await web3.eth.getBalance(accounts[i]);
                balances.push(web3.utils.toBN(balance));
            }

            // account 2 claims as account 1
            let delegate = web3.utils.fromAscii("cobo");
            await this.hermes.distributeRewards(delegate, 1600, [accounts[1], accounts[3]], [100000000, 200000000], {from: accounts[0], value: MinTips + 300000000});
            assert.equal(await web3.eth.getBalance(accounts[2]), balances[2].add(web3.utils.toBN(100000000)).toString());
            assert.equal(await web3.eth.getBalance(accounts[3]), balances[3].add(web3.utils.toBN(200000000)).toString());

            // invalid signature
            let err;
            try {
                await this.contract.registerForwardService(accounts[1], registerNonce+1, 1500, sig, {from: accounts[3]});
            } catch (e) {
                err = e;
            }
            assert.ok(err.toString().includes("invalid signature"))

            // invalid nonce
            registerMsgHash = registerNonce + 'I authorize ' + accounts[3].toLowerCase() + ' to register in ' + this.contract.address.toLowerCase();
            sig = await web3.eth.accounts.sign(registerMsgHash, '0xae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f');
            sig = sig.signature;
            try {
                await this.contract.registerForwardService(accounts[1], registerNonce, 1500, sig, {from: accounts[3]});
            } catch (e) {
                err = e;
            }
            assert.ok(err.toString().includes("nonce is invalid"))

            // succeed to deregister
            const deregisterNonce = 3;
            const deregisterMsgHash = deregisterNonce + 'I authorize ' + accounts[2].toLowerCase() + ' to deregister in ' + this.contract.address.toLowerCase();
            sig = await web3.eth.accounts.sign(deregisterMsgHash, '0xae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f');
            sig = sig.signature;
            await this.contract.deregisterForwardService(accounts[1], deregisterNonce, sig, {from: accounts[2]});

            // account 1 claims for self
            await this.hermes.distributeRewards(delegate, 1700, [accounts[1]], [100000000], {from: accounts[0], value: MinTips + 100000000});
            assert.equal(await web3.eth.getBalance(accounts[1]), balances[1].add(web3.utils.toBN(100000000)).toString());
        });
    });
});