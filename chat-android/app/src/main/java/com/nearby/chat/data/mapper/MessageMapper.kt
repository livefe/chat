package com.nearby.chat.data.mapper

import com.nearby.chat.core.model.ConversationItem
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.NearbyBootstrap
import com.nearby.chat.core.model.NearbyRoomItem
import com.nearby.chat.core.model.RoomApplicationItem
import com.nearby.chat.core.model.RoomBootstrap
import com.nearby.chat.core.model.RoomDetail
import com.nearby.chat.data.remote.dto.ConversationItemDto
import com.nearby.chat.data.remote.dto.MessageItemDto
import com.nearby.chat.data.remote.dto.NearbyBootstrapResultDto
import com.nearby.chat.data.remote.dto.NearbyRoomItemDto
import com.nearby.chat.data.remote.dto.RoomApplicationItemDto
import com.nearby.chat.data.remote.dto.RoomBootstrapResultDto
import com.nearby.chat.data.remote.dto.RoomDetailDto

fun MessageItemDto.toDomain(): MessageItem = MessageItem(
    id = id,
    scope = scope,
    roomId = roomId,
    postId = postId,
    fromUserId = fromUserId,
    toUserId = toUserId,
    fromName = fromName,
    fromAvatar = fromAvatar,
    contentType = contentType,
    content = content,
    replyToId = replyToId,
    replyToFromName = replyToFromName,
    replyToContentType = replyToContentType,
    replyToContent = replyToContent,
    latitude = latitude,
    longitude = longitude,
    createdAt = createdAt,
    recalled = recalled,
)

fun ConversationItemDto.toDomain(): ConversationItem = ConversationItem(
    peerId = peerId,
    peerNickname = peerNickname,
    peerAvatar = peerAvatar,
    peerMobile = peerMobile,
    lastContent = lastContent,
    lastActiveAt = lastActiveAt,
    unreadCount = unreadCount,
    pinned = pinned,
)

fun NearbyRoomItemDto.toDomain(): NearbyRoomItem = NearbyRoomItem(
    id = id,
    name = name,
    placeType = placeType,
    placeTypeText = placeTypeText,
    shapeType = shapeType,
    centerLat = centerLat,
    centerLng = centerLng,
    radius = radius,
    distance = distance,
    requireGeofence = requireGeofence,
)

fun NearbyBootstrapResultDto.toDomain(): NearbyBootstrap = NearbyBootstrap(
    rooms = rooms.map { it.toDomain() },
    onlineCount = onlineCount,
    messages = messages.map { it.toDomain() },
)

fun RoomDetailDto.toDomain(): RoomDetail = RoomDetail(
    id = id,
    name = name,
    description = description,
    placeType = placeType,
    placeTypeText = placeTypeText,
    shapeType = shapeType,
    centerLat = centerLat,
    centerLng = centerLng,
    radius = radius,
    requireGeofence = requireGeofence,
)

fun RoomBootstrapResultDto.toDomain(): RoomBootstrap = RoomBootstrap(
    room = room.toDomain(),
    onlineCount = onlineCount,
    messages = messages.map { it.toDomain() },
)

fun RoomApplicationItemDto.toDomain(): RoomApplicationItem = RoomApplicationItem(
    id = id,
    name = name,
    description = description,
    placeType = placeType,
    placeTypeText = placeTypeText,
    status = status,
    statusText = statusText,
    radius = radius,
    createdAt = createdAt,
)
