package com.nearby.chat.core.model

data class PostExtras(
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
