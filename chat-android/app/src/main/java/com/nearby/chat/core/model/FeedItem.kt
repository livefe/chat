package com.nearby.chat.core.model

sealed class FeedItem {
    data class Post(
        val post: PostItem,
    ) : FeedItem()

    data class Sponsored(
        val sponsored: SponsoredItem,
    ) : FeedItem()
}
