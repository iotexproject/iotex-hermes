pragma solidity ^0.4.24;

import '../ownership/Ownable.sol';

/**
 * @title SafeMath
 * @dev Math operations with safety checks that throw on error
 */
library SafeMath {
  function mul(uint256 a, uint256 b) internal constant returns (uint256) {
    uint256 c = a * b;
    assert(a == 0 || c / a == b);
    return c;
  }

  function div(uint256 a, uint256 b) internal constant returns (uint256) {
    // assert(b > 0); // Solidity automatically throws when dividing by 0
    uint256 c = a / b;
    // assert(a == b * c + a % b); // There is no case in which this doesn't hold
    return c;
  }

  function sub(uint256 a, uint256 b) internal constant returns (uint256) {
    assert(b <= a);
    return a - b;
  }

  function add(uint256 a, uint256 b) internal constant returns (uint256) {
    uint256 c = a + b;
    assert(c >= a);
    return c;
  }
}

contract MultiSendMock is Ownable {
    using SafeMath for uint256;

    uint256 public minTips;
    uint256 public limit;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Receipt(address _token, uint256 _totalAmount, uint256 _tips, string _payload);

    constructor(uint256 _minTips, uint256 _limit) public {
        minTips = _minTips;
        limit = _limit;
    }

    function sendCoin(address[] recipients, uint256[] amounts, string payload) public payable {
        require(recipients.length == amounts.length);
        require(recipients.length <= limit);
        uint256 totalAmount = minTips;
        for (uint256 i = 0; i < recipients.length; i++) {
            totalAmount = totalAmount.add(amounts[i]);
            require(msg.value >= totalAmount);
            recipients[i].transfer(amounts[i]);
            emit Transfer(msg.sender, recipients[i], amounts[i]);
        }
        emit Receipt(address(this), totalAmount.sub(minTips), minTips.add(msg.value.sub(totalAmount)), payload);
    }
}



