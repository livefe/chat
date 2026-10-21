package com.nearby.chat.core.model

data class PostSignupListCounts(
    val total: Int = 0,
    val confirmed: Int = 0,
    val waitlisted: Int = 0,
    val cancelled: Int = 0,
    val notCheckedIn: Int = 0,
    val checkedIn: Int = 0,
)
