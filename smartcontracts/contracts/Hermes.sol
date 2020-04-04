pragma solidity ^0.4.24;

import './math/SafeMath.sol';
import './ownership/Whitelist.sol';
import './ForwardRegistration.sol';

contract Multisend is Ownable {
    function sendCoin(address[] recipients, uint256[] amounts, string payload) public payable;
    function minTips() public view returns (uint256);
}

contract Hermes is Whitelist {
    using SafeMath for uint256;

    struct Distribution{
        uint256 distributedCount;
        uint256 amount;
    }

    uint256 public contractStartEpoch;
    uint256[] public endEpochs;
    mapping(bytes32 => mapping(address => uint256)) public recipientEpochTracker;
    mapping(bytes32 => mapping(uint256 => Distribution)) public distributions;
    mapping(bytes32 => uint256) public distributedCount;
    mapping(bytes32 => uint256) public distributedAmount;
    string public analyticsEndpoint;

    Multisend public multisender;
    ForwardRegistration public forwardRegistration;

    event Distribute(uint256 startEpoch, uint256 endEpoch, bytes32 indexed delegateName, uint256 numOfRecipients, uint256 totalAmount);
    event CommitDistributions(uint256 endEpoch, bytes32[] delegateNames);

    constructor(
        uint256 _contractStartEpoch,
        address _multisendAddress,
        address _forwardRegistrationAddress,
        string _analyticsEndpoint
    ) public {
        addAddressToWhitelist(msg.sender);
        contractStartEpoch = _contractStartEpoch;
        multisender = Multisend(_multisendAddress);
        forwardRegistration = ForwardRegistration(_forwardRegistrationAddress);
        analyticsEndpoint = _analyticsEndpoint;
    }

    function getEndEpochCount() public view returns (uint256) {
        return endEpochs.length;
    }

    function setMultisendAddress(address _multisendAddress) public onlyWhitelisted {
        multisender = Multisend(_multisendAddress);
    }

    function setAnalyticsEndpoint(string _endpoint) public onlyWhitelisted {
        analyticsEndpoint = _endpoint;
    }

    function distributeRewards(
        bytes32 delegateName,
        uint256 endEpoch,
        address[] recipients,
        uint256[] amounts
    ) public payable onlyWhitelisted {
        uint256 lastEpoch = 0;
        if (endEpochs.length > 0) {
            lastEpoch = endEpochs[endEpochs.length-1];
        }
        require(endEpoch > lastEpoch, "invalid end epoch");
        require(recipients.length > 0 && recipients.length == amounts.length, "invalid number of recipients");
        uint256 startEpoch = lastEpoch + 1;

        uint256 totalAmount;
        for (uint256 i = 0; i < recipients.length; i++) {
            address recipient = recipients[i];
            uint256 curEpoch = recipientEpochTracker[delegateName][recipient];
            require(curEpoch < endEpoch, "reward has already been distributed to the recipient");
            updateRecipientTracker(endEpoch, delegateName, recipient);
            totalAmount = totalAmount.add(amounts[i]);
            recipients[i] = forwardRegistration.getForwardAddress(recipient, endEpoch);
        }

        uint256 minTips = multisender.minTips();

        require(msg.value == totalAmount.add(minTips), "message value does not match the sending amount");
        updateDistributionProgress(delegateName, recipients.length, totalAmount);

        string memory payload = getPayload(delegateName, startEpoch, endEpoch);

        multisender.sendCoin.value(msg.value)(recipients, amounts, payload);

        emit Distribute(startEpoch, endEpoch, delegateName, recipients.length, totalAmount);
    }

    function commitDistributions(uint256 endEpoch, bytes32[] delegateNames) public onlyWhitelisted {
        for (uint256 i = 0; i < delegateNames.length; i++) {
            bytes32 delegateName = delegateNames[i];
            distributions[delegateName][endEpoch] = Distribution(distributedCount[delegateName], distributedAmount[delegateName]);
            distributedCount[delegateName] = 0;
            distributedAmount[delegateName] = 0;
        }
        endEpochs.push(endEpoch);

        emit CommitDistributions(endEpoch, delegateNames);
    }

    function updateRecipientTracker(uint256 endEpoch, bytes32 delegateName, address recipient) internal {
        recipientEpochTracker[delegateName][recipient] = endEpoch;
    }

    function updateDistributionProgress(bytes32 delegateName, uint256 numberOfRecipients, uint256 totalAmount) internal {
        distributedCount[delegateName] += numberOfRecipients;
        distributedAmount[delegateName] = distributedAmount[delegateName].add(totalAmount);
    }

    function getPayload(bytes32 delegateName, uint256 startEpoch, uint256 endEpoch) internal view returns (string) {
        return string(abi.encodePacked(
            "Reward distribution for ", bytes32ToString(delegateName), " from epoch ", uint2str(startEpoch), " to epoch ", uint2str(endEpoch), ". Please see ", analyticsEndpoint, " for reward details."
        ));
    }

    function uint2str(uint i) internal pure returns (string) {
        if (i == 0) {
            return "0";
        }
        uint j = i;
        uint length;
        while (j != 0){
            length++;
            j /= 10;
        }
        bytes memory b = new bytes(length);
        uint k = length - 1;
        while (i != 0){
            b[k--] = byte(48 + i % 10);
            i /= 10;
        }
        return string(b);
    }

    function bytes32ToString(bytes32 _bytes32) internal pure returns (string) {
        bytes memory bytesArray = new bytes(32);
        for (uint256 i; i < 32; i++) {
            bytesArray[i] = _bytes32[i];
        }
        return string(bytesArray);
    }
}