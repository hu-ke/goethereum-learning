//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

contract MyNFT is ERC721URIStorage {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;

    struct NFT {
        uint256 price;
        bool forSale;
    }

    mapping(uint256 => NFT) public nfts;
    mapping(string => uint256) private _tokenIdsByURI;
    mapping(address => uint256[]) private _ownedTokens;
    mapping(uint256 => uint256) private _ownedTokensIndex;

    constructor() ERC721("MyNFT", "MNFT") {}

    function setForSale(uint256 tokenId, uint256 price) internal {
        require(ownerOf(tokenId) == msg.sender, "You are not the owner");
        nfts[tokenId] = NFT(price, true);
    }

    function mint(string memory tokenURI, uint256 price) public returns (uint256) {
        _tokenIds.increment();

        uint256 newItemId = _tokenIds.current();
        // _mint 方法通常用于创建新的代币并将其分配给指定的账户。这个方法在各种代币标准（如 ERC-20 和 ERC-721）中都有广泛的应用。
        _mint(msg.sender, newItemId);
        _setTokenURI(newItemId, tokenURI);
        _tokenIdsByURI[tokenURI] = newItemId;
        _addTokenToOwnerEnumeration(msg.sender, newItemId);
        setForSale(newItemId, price);

        return newItemId;
    }

    function buyNFT(uint256 tokenId) external payable {
        NFT storage nft = nfts[tokenId];
        require(nft.forSale, "NFT is not for sale");
        require(msg.value == nft.price, "Incorrect value sent");

        address seller = ownerOf(tokenId);

        // Transfer the NFT to the buyer
        _transfer(seller, msg.sender, tokenId);

        // Transfer the ETH to the seller
        payable(seller).transfer(msg.value);

        // Update the NFT status
        nft.forSale = false;
    }

    function getTokenIdByTokenURI(string memory tokenURI) public view returns (uint256) {
        require(_tokenIdsByURI[tokenURI] != 0, "Token URI does not exist");
        return _tokenIdsByURI[tokenURI];
    }

    function _addTokenToOwnerEnumeration(address to, uint256 tokenId) private {
        _ownedTokensIndex[tokenId] = _ownedTokens[to].length;
        _ownedTokens[to].push(tokenId);
    }

    function _removeTokenFromOwnerEnumeration(address from, uint256 tokenId) private {
        uint256 lastTokenIndex = _ownedTokens[from].length - 1;
        uint256 tokenIndex = _ownedTokensIndex[tokenId];

        if (tokenIndex != lastTokenIndex) {
            uint256 lastTokenId = _ownedTokens[from][lastTokenIndex];
            _ownedTokens[from][tokenIndex] = lastTokenId;
            _ownedTokensIndex[lastTokenId] = tokenIndex;
        }

        _ownedTokens[from].pop();
        delete _ownedTokensIndex[tokenId];
    }

    function _beforeTokenTransfer(address from, address to, uint256 tokenId, uint256 batchSize) internal virtual override {
        super._beforeTokenTransfer(from, to, tokenId, batchSize);

        if (from != address(0)) {
            _removeTokenFromOwnerEnumeration(from, tokenId);
        }
        if (to != address(0)) {
            _addTokenToOwnerEnumeration(to, tokenId);
        }
    }

    function getTokensOfOwner(address owner) public view returns (uint256[] memory) {
        return _ownedTokens[owner];
    }

    function totalSupply() public view returns (uint256) {
        return _tokenIds.current();
    }
}
