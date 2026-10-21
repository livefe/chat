package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostExtrasDto(
    val amount: Double = 0.0,
    val negotiable: Boolean = false,
    val needSignup: Boolean = false,
    val needCheckIn: Boolean = false,
    val maxParticipants: Int = 0,
    val allowWaitlist: Boolean = false,
    val signupDeadlineAt: Long = 0,
    val feeType: Int = 0,
    val feeDescription: String = "",
)
