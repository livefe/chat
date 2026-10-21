package com.nearby.chat.feature.post

import com.nearby.chat.core.model.PostItem

internal sealed class MyPostConfirm {
    abstract val post: PostItem

    data class Withdraw(override val post: PostItem) : MyPostConfirm()
    data class DeletePublished(override val post: PostItem) : MyPostConfirm()
    data class DeleteJoined(override val post: PostItem) : MyPostConfirm()
}
