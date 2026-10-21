package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SupportSessionDto(
    val id: Long = 0,
    val status: Int = 0,
    val statusText: String = "",
    val lastMessageId: Long = 0,
    val lastContent: String = "",
    val lastActiveAt: Long = 0,
    val unreadCount: Long = 0,
    val assigneeEmployeeId: Long = 0,
    val assigneeName: String = "",
    val queueAt: Long = 0,
    val brandName: String = "",
    val brandAvatarUrl: String = "",
)
