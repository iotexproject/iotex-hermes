const MinTips = 1000;
const Limit = 100;
const ContractStartEpoch = 1000;
const AnalyticsEndpoint = "https://analytics.iotexscan.io/";

const Migrations = artifacts.require("Migrations");
const ForwardRegistration = artifacts.require("ForwardRegistration");
const Hermes = artifacts.require("Hermes");
const MultiSendMock = artifacts.require("MultiSendMock");

module.exports = function(deployer) {
  deployer.deploy(Migrations);
  deployer.deploy(ForwardRegistration);
  deployer.deploy(MultiSendMock, MinTips, Limit).then(function() {
    return deployer.deploy(Hermes, ContractStartEpoch, MultiSendMock.address, ForwardRegistration.address, AnalyticsEndpoint);
  });
};
