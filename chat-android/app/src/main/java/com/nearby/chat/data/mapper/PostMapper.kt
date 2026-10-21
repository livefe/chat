package com.nearby.chat.data.mapper

import com.nearby.chat.core.model.AdEvent
import com.nearby.chat.core.model.FeedItem
import com.nearby.chat.core.model.PostCheckInCode
import com.nearby.chat.core.model.PostCheckInResult
import com.nearby.chat.core.model.PostDiscussPreview
import com.nearby.chat.core.model.PostExtras
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.model.PostPublisher
import com.nearby.chat.core.model.PostSignupItem
import com.nearby.chat.core.model.PostSignupListCounts
import com.nearby.chat.core.model.PostSignupPage
import com.nearby.chat.core.model.PostSignupSummary
import com.nearby.chat.core.model.SponsoredItem
import com.nearby.chat.data.remote.dto.AdEventItemRequest
import com.nearby.chat.data.remote.dto.FeedItemDto
import com.nearby.chat.data.remote.dto.PostCheckInCodeDto
import com.nearby.chat.data.remote.dto.PostCheckInResultDto
import com.nearby.chat.data.remote.dto.PostDiscussPreviewDto
import com.nearby.chat.data.remote.dto.PostExtrasDto
import com.nearby.chat.data.remote.dto.PostItemDto
import com.nearby.chat.data.remote.dto.PostPublisherDto
import com.nearby.chat.data.remote.dto.PostSignupItemDto
import com.nearby.chat.data.remote.dto.PostSignupListCountsDto
import com.nearby.chat.data.remote.dto.PostSignupListResultDto
import com.nearby.chat.data.remote.dto.PostSignupSummaryDto
import com.nearby.chat.data.remote.dto.SponsoredItemDto

fun PostExtrasDto.toDomain(): PostExtras = PostExtras(
    amount = amount,
    negotiable = negotiable,
    needSignup = needSignup,
    needCheckIn = needCheckIn,
    maxParticipants = maxParticipants,
    allowWaitlist = allowWaitlist,
    signupDeadlineAt = signupDeadlineAt,
    feeType = feeType,
    feeDescription = feeDescription,
)

fun PostExtras.toDto(): PostExtrasDto = PostExtrasDto(
    amount = amount,
    negotiable = negotiable,
    needSignup = needSignup,
    needCheckIn = needCheckIn,
    maxParticipants = maxParticipants,
    allowWaitlist = allowWaitlist,
    signupDeadlineAt = signupDeadlineAt,
    feeType = feeType,
    feeDescription = feeDescription,
)

fun PostSignupSummaryDto.toDomain(): PostSignupSummary = PostSignupSummary(
    confirmed = confirmed,
    waitlisted = waitlisted,
)

fun PostDiscussPreviewDto.toDomain(): PostDiscussPreview = PostDiscussPreview(
    fromName = fromName,
    contentType = contentType,
    content = content,
)

fun PostPublisherDto.toDomain(): PostPublisher = PostPublisher(
    id = id,
    nickname = nickname,
    avatarUrl = avatarUrl,
)

fun PostItemDto.toDomain(): PostItem = PostItem(
    id = id,
    roomId = roomId,
    category = category,
    pinned = pinned,
    title = title,
    content = content,
    images = images,
    imageIds = imageIds,
    placeName = placeName,
    latitude = latitude,
    longitude = longitude,
    startAt = startAt,
    endAt = endAt,
    status = status,
    publisherId = publisherId,
    publisher = publisher?.toDomain(),
    discussMessageCount = discussMessageCount,
    discussPreview = discussPreview?.toDomain(),
    createdAt = createdAt,
    updatedAt = updatedAt,
    distance = distance,
    extras = extras?.toDomain(),
    signupSummary = signupSummary?.toDomain(),
    mySignupStatus = mySignupStatus,
    myCheckedInAt = myCheckedInAt,
)

fun SponsoredItemDto.toDomain(): SponsoredItem = SponsoredItem(
    campaignId = campaignId,
    title = title,
    subtitle = subtitle,
    imageUrl = imageUrl,
    advertiserName = advertiserName,
    badge = badge,
    landingType = landingType,
    landingValue = landingValue,
    distance = distance,
    impressionToken = impressionToken,
    requestId = requestId,
    slotIndex = slotIndex,
)

fun FeedItemDto.toDomain(): FeedItem? = when (type) {
    "sponsored" -> sponsored?.let { FeedItem.Sponsored(it.toDomain()) }
    else -> post?.let { FeedItem.Post(it.toDomain()) }
}

fun AdEvent.toRequest(): AdEventItemRequest = AdEventItemRequest(
    requestId = requestId,
    campaignId = campaignId,
    eventType = eventType,
    impressionToken = impressionToken,
    slotIndex = slotIndex,
    latitude = latitude,
    longitude = longitude,
)

fun PostCheckInCodeDto.toDomain(): PostCheckInCode = PostCheckInCode(
    postId = postId,
    postTitle = postTitle,
    placeName = placeName,
    startAt = startAt,
    endAt = endAt,
    payload = payload,
    checkedIn = checkedIn,
    checkedInAt = checkedInAt,
    checkedInAtText = checkedInAtText,
    userNickname = userNickname,
    signupAt = signupAt,
)

fun PostCheckInResultDto.toDomain(): PostCheckInResult = PostCheckInResult(
    postId = postId,
    postTitle = postTitle,
    userId = userId,
    userNickname = userNickname,
    userAvatarUrl = userAvatarUrl,
    signupAt = signupAt,
    checkedInAt = checkedInAt,
    checkedInAtText = checkedInAtText,
    alreadyCheckedIn = alreadyCheckedIn,
    operatorId = operatorId,
    operatorNickname = operatorNickname,
    operatorAvatarUrl = operatorAvatarUrl,
)

fun PostSignupItemDto.toDomain(): PostSignupItem = PostSignupItem(
    id = id,
    userId = userId,
    userNickname = userNickname,
    userMobile = userMobile,
    signupStatus = signupStatus,
    signupStatusText = signupStatusText,
    checkedInAt = checkedInAt,
    createdAt = createdAt,
    cancelledAt = cancelledAt,
    remark = remark,
)

fun PostSignupListCountsDto.toDomain(): PostSignupListCounts = PostSignupListCounts(
    total = total,
    confirmed = confirmed,
    waitlisted = waitlisted,
    cancelled = cancelled,
    notCheckedIn = notCheckedIn,
    checkedIn = checkedIn,
)

fun PostSignupListResultDto.toDomain(): PostSignupPage = PostSignupPage(
    total = total,
    list = list.map { it.toDomain() },
    needCheckIn = needCheckIn,
    counts = counts.toDomain(),
)
