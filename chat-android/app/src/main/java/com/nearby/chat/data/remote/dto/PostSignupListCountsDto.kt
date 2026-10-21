package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class PostSignupListCountsDto(
    val total: Int = 0,
    val confirmed: Int = 0,
    val waitlisted: Int = 0,
    val cancelled: Int = 0,
    val notCheckedIn: Int = 0,
    val checkedIn: Int = 0,
)
