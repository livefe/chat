package com.nearby.chat.feature.account

data class EditNicknameUiState(
    val original: String = "",
    val draft: String = "",
    val saving: Boolean = false,
) {
    val canSave: Boolean
        get() = draft.trim().isNotEmpty() && draft.trim() != original.trim() && !saving
}
