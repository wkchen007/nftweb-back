// SPDX-License-Identifier: MIT
pragma solidity >=0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";

contract MyNFT is ERC721 {
    using Strings for uint256;

    address owner;
    uint256 public maxSupply = 9; // 最大發行量
    bool private isOpened = false; //盲盒是否打開

    uint256 public counter = 0;

    modifier onlyOwner() {
        require(msg.sender == owner);
        _;
    }

    constructor(
        string memory _name,
        string memory _symbol
    ) ERC721(_name, _symbol) {
        owner = msg.sender;
    }

    //開盲盒
    function openBlindBox() external onlyOwner {
        isOpened = true;
    }

    //設定NFT的baseURI(盲盒)
    function _baseURI() internal pure override returns (string memory) {
        return
            "ipfs://bafkreicnovsrbhko6exqtctuhqyg6nvloulmydgu4onfzpp4uqkm7hxle4/";
    }

    //查看NFT Metadata網址
    function tokenURI(
        uint256 tokenId
    ) public view override returns (string memory) {
        if (!isOpened) {
            return _baseURI();
        }
        return
            string(
                abi.encodePacked(
                    "ipfs://bafybeiectyzmhyq5mjqvdbk3x2g77zbs7odwdkfqzuyzfpey3bxweirasu/",
                    tokenId.toString(),
                    ".json"
                )
            );
    }

    // 實作mint function，主要用來 demo 確認用
    function mint(address to, uint256 amount) external payable {
        uint256 price = 0.01 ether;
        require(counter + amount <= maxSupply, "over max supply.");
        require(msg.value == price * amount, "incorrect payment");

        for (uint256 i = 0; i < amount; i++) {
            _mint(to, counter);
            unchecked {
                counter++;
            }
        }
    }

    // 提領合約餘額（只有 owner 可以呼叫）
    function withdraw() external onlyOwner {
        uint256 balance = address(this).balance;
        require(balance > 0, "no funds");

        (bool success, ) = payable(owner).call{value: balance}("");
        require(success, "withdraw failed");
    }
}
