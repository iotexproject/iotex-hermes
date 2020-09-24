pragma solidity ^0.4.24;

import "./utils/Pausable.sol";
import "./ownership/Ownable.sol";

contract AutoDepositRegister is Pausable, Ownable {
	mapping (address => int256) public buckets;
	mapping (address => bool) public registrants;

	function pause() public onlyOwner {
		_pause();
	}

	function unpause() public onlyOwner {
		_unpause();
	}

	function register(int256 bucketId) public whenNotPaused {
		registrants[msg.sender] = true;
		buckets[msg.sender] = bucketId;
	}

	function unregister() public whenNotPaused {
		registrants[msg.sender] = false;
		buckets[msg.sender] = -1;
	}

	function bucket(address owner) public view returns (int256) {
		if (registrants[owner]) {
			return buckets[owner];
		}
		return -1;
	}
}
