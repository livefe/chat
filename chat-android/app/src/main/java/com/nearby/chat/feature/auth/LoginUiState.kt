package com.nearby.chat.feature.auth

data class LoginUiState(
    val mobile: String = "",
    val code: String = "",
    val agreed: Boolean = false,
    val loading: Boolean = false,
    val sendingCode: Boolean = false,
    val countdown: Int = 0,
)
