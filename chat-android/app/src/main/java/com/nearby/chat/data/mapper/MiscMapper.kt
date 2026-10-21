package com.nearby.chat.data.mapper

import com.nearby.chat.core.model.AppUpdateInfo
import com.nearby.chat.core.model.FeedbackItem
import com.nearby.chat.core.model.RegionItem
import com.nearby.chat.core.model.UploadedMedia
import com.nearby.chat.data.remote.dto.AppUpdateCheckDto
import com.nearby.chat.data.remote.dto.FeedbackItemDto
import com.nearby.chat.data.remote.dto.RegionItemDto
import com.nearby.chat.data.remote.dto.UploadMediaDto

fun FeedbackItemDto.toDomain(): FeedbackItem = FeedbackItem(
    id = id,
    type = type,
    typeText = typeText,
    content = content,
    contact = contact,
    images = images,
    status = status,
    statusText = statusText,
    reply = reply,
    handledAt = handledAt,
    createdAt = createdAt,
)

fun AppUpdateCheckDto.toDomain(): AppUpdateInfo = AppUpdateInfo(
    latestVersion = latestVersion,
    minSupportedVersion = minSupportedVersion,
    downloadUrl = downloadUrl,
    fileSize = fileSize,
    hasUpdate = hasUpdate,
    forceUpdate = forceUpdate,
    title = title,
    message = message,
)

fun RegionItemDto.toDomain(): RegionItem = RegionItem(
    code = code,
    name = name,
    shortName = shortName,
    parentCode = parentCode,
    level = level,
    fullPath = fullPath,
    tiandituGb = tiandituGb,
    amapAdcode = amapAdcode,
)

fun UploadMediaDto.toDomain(): UploadedMedia = UploadedMedia(
    id = id,
    url = url,
)
