package com.nearby.chat.feature.account

data class AccountDeactivateUiState(
    val mobile: String = "",
    val code: String = "",
    val agreed: Boolean = false,
    val counting: Int = 0,
    val sending: Boolean = false,
    val submitting: Boolean = false,
    val confirmVisible: Boolean = false,
)
