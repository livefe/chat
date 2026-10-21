package com.nearby.chat.data.mapper

import com.nearby.chat.core.model.SupportMessage
import com.nearby.chat.core.model.SupportProfile
import com.nearby.chat.core.model.SupportSession
import com.nearby.chat.data.remote.dto.SupportMessageDto
import com.nearby.chat.data.remote.dto.SupportProfileDto
import com.nearby.chat.data.remote.dto.SupportSessionDto

fun SupportProfileDto.toDomain(): SupportProfile = SupportProfile(
    brandName = brandName,
    brandAvatarUrl = brandAvatarUrl,
    welcomeText = welcomeText,
    offlineText = offlineText,
    enabled = enabled,
)

fun SupportSessionDto.toDomain(): SupportSession = SupportSession(
    id = id,
    status = status,
    statusText = statusText,
    lastMessageId = lastMessageId,
    lastContent = lastContent,
    lastActiveAt = lastActiveAt,
    unreadCount = unreadCount,
    assigneeEmployeeId = assigneeEmployeeId,
    assigneeName = assigneeName,
    queueAt = queueAt,
    brandName = brandName,
    brandAvatarUrl = brandAvatarUrl,
)

fun SupportMessageDto.toDomain(): SupportMessage = SupportMessage(
    id = id,
    conversationId = conversationId,
    senderKind = senderKind,
    senderId = senderId,
    senderName = senderName,
    senderAvatarUrl = senderAvatarUrl,
    contentType = contentType,
    content = content,
    createdAt = createdAt,
    recalled = recalled,
    clientMsgId = clientMsgId,
)
