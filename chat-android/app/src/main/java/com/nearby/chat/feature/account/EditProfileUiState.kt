package com.nearby.chat.feature.account

data class EditProfileUiState(
    val nickname: String = "",
    val mobile: String = "",
    val avatarUrl: String = "",
    val localAvatarUri: String = "",
    val uploadingAvatar: Boolean = false,
    val saving: Boolean = false,
)
