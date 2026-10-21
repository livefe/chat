package com.nearby.chat.core.post

import com.nearby.chat.core.model.CheckInResultPayload
import com.nearby.chat.core.model.PostCheckInResult
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

private val checkInResultJson = Json { ignoreUnknownKeys = true }

@Serializable
private data class CheckInResultPayloadDto(
    val kind: String = "",
    val message: String = "",
    val postId: Long = 0,
    val postTitle: String = "",
    val userId: Long = 0,
    val userNickname: String = "",
    val userAvatarUrl: String = "",
    val signupAt: String = "",
    val checkedInAtText: String = "",
    val operatorNickname: String = "",
    val operatorAvatarUrl: String = "",
)

private fun CheckInResultPayload.toDto() = CheckInResultPayloadDto(
    kind = kind,
    message = message,
    postId = postId,
    postTitle = postTitle,
    userId = userId,
    userNickname = userNickname,
    userAvatarUrl = userAvatarUrl,
    signupAt = signupAt,
    checkedInAtText = checkedInAtText,
    operatorNickname = operatorNickname,
    operatorAvatarUrl = operatorAvatarUrl,
)

private fun CheckInResultPayloadDto.toDomain() = CheckInResultPayload(
    kind = kind,
    message = message,
    postId = postId,
    postTitle = postTitle,
    userId = userId,
    userNickname = userNickname,
    userAvatarUrl = userAvatarUrl,
    signupAt = signupAt,
    checkedInAtText = checkedInAtText,
    operatorNickname = operatorNickname,
    operatorAvatarUrl = operatorAvatarUrl,
)

fun buildCheckInResultPayload(
    kind: String,
    result: PostCheckInResult? = null,
    message: String = "",
): CheckInResultPayload {
    if (kind == "fail") {
        return CheckInResultPayload(kind = kind, message = message)
    }
    val data = result ?: return CheckInResultPayload(kind = kind)
    return CheckInResultPayload(
        kind = kind,
        postId = data.postId,
        postTitle = data.postTitle,
        userId = data.userId,
        userNickname = data.userNickname,
        userAvatarUrl = data.userAvatarUrl,
        signupAt = data.signupAt,
        checkedInAtText = data.checkedInAtText,
        operatorNickname = data.operatorNickname,
        operatorAvatarUrl = data.operatorAvatarUrl,
    )
}

fun encodeCheckInResultPayload(payload: CheckInResultPayload): String = checkInResultJson.encodeToString(payload.toDto())

fun decodeCheckInResultPayload(raw: String): CheckInResultPayload? = runCatching { checkInResultJson.decodeFromString<CheckInResultPayloadDto>(raw).toDomain() }.getOrNull()
