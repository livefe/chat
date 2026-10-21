package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class SubmitReportRequest(
    val targetType: Int,
    val targetId: Long,
    val reason: Int,
    val content: String? = null,
    val images: List<Long>? = null,
)
