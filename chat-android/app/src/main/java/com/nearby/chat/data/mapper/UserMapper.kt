package com.nearby.chat.data.mapper

import com.nearby.chat.core.model.BlockListItem
import com.nearby.chat.core.model.PublicUserProfile
import com.nearby.chat.core.model.UserInfo
import com.nearby.chat.data.remote.dto.BlockListItemDto
import com.nearby.chat.data.remote.dto.PublicUserProfileDto
import com.nearby.chat.data.remote.dto.UserInfoDto

fun UserInfoDto.toDomain(): UserInfo = UserInfo(
    id = id,
    mobile = mobile,
    nickname = nickname,
    avatarUrl = avatarUrl,
    status = status,
)

fun UserInfo.toDto(): UserInfoDto = UserInfoDto(
    id = id,
    mobile = mobile,
    nickname = nickname,
    avatarUrl = avatarUrl,
    status = status,
)

fun PublicUserProfileDto.toDomain(): PublicUserProfile = PublicUserProfile(
    id = id,
    nickname = nickname,
    avatarUrl = avatarUrl,
    createdAt = createdAt,
    isSelf = isSelf,
    deactivated = deactivated,
    blockedByMe = blockedByMe,
    canMessage = canMessage,
)

fun BlockListItemDto.toDomain(): BlockListItem = BlockListItem(
    userId = userId,
    nickname = nickname,
    avatarUrl = avatarUrl,
    createdAt = createdAt,
)
