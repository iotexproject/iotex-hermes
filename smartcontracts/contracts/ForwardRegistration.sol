pragma solidity ^0.4.24;

contract ForwardRegistration {
    struct Forward {
        uint256 nonce;
        address destination;
        uint256 startEpoch;
    }

    mapping(address => Forward) public forwardService;

    event RegisterForwardService(address indexed owner, address indexed alternative, uint256 epoch);
    event DeregisterForwardService(address indexed owner);

    function registerForwardService(address owner, uint256 nonce, uint256 startEpoch, bytes signature) public {
        checkSignAndNonce(
            owner,
            nonce,
            abi.encodePacked(uint2str(nonce), "I authorize", addrToString(msg.sender), " to register in", addrToString(address(this))),
            signature
        );
        forwardService[owner] = Forward(nonce, msg.sender, startEpoch);
        emit RegisterForwardService(owner, msg.sender, startEpoch);
    }

    function deregisterForwardService(address owner, uint256 nonce, bytes signature) public {
        checkSignAndNonce(
            owner,
            nonce,
            abi.encodePacked(uint2str(nonce), "I authorize", addrToString(msg.sender), " to deregister in", addrToString(address(this))),
            signature
        );
        forwardService[owner] = Forward(nonce, address(0), 0);
        emit DeregisterForwardService(owner);
    }

    function getForwardAddress(address owner, uint256 endEpoch) public view returns (address) {
        if (forwardService[owner].destination != address(0) && endEpoch >= forwardService[owner].startEpoch) {
            return forwardService[owner].destination;
        }
        return owner;
    }

    function checkSignAndNonce(address owner, uint256 nonce, bytes message, bytes signature) internal view {
        require(recover(toEthPersonalSignedMessageHash(message), signature) == owner, "invalid signature");
        require(nonce > forwardService[owner].nonce, "nonce is invalid");
    }

    function recover(bytes32 hash, bytes signature)
        internal
        pure
        returns (address)
    {
        bytes32 r;
        bytes32 s;
        uint8 v;
        // Check the signature length
        if (signature.length != 65) {
            return (address(0));
        }
        // Divide the signature in r, s and v variables with inline assembly.
        assembly {
            r := mload(add(signature, 0x20))
            s := mload(add(signature, 0x40))
            v := byte(0, mload(add(signature, 0x60)))
        }
        // Version of signature should be 27 or 28, but 0 and 1 are also possible versions
        if (v < 27) {
            v += 27;
        }
        // If the version is correct return the signer address
        if (v != 27 && v != 28) {
            return (address(0));
        }
        return ecrecover(hash, v, r, s);
    }

    function toEthPersonalSignedMessageHash(bytes _msg) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n", uint2str(_msg.length), _msg));
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

    function addrToString(address _addr) internal pure returns(string) {
        bytes32 value = bytes32(uint256(_addr));
        bytes memory alphabet = "0123456789abcdef";

        bytes memory str = new bytes(43);
        str[0] = ' ';
        str[1] = '0';
        str[2] = 'x';
        for (uint i = 0; i < 20; i++) {
            str[3+i*2] = alphabet[uint(value[i + 12] >> 4)];
            str[4+i*2] = alphabet[uint(value[i + 12] & 0x0f)];
        }
        return string(str);
    }
}