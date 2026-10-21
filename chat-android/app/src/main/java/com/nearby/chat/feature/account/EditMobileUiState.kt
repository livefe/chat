package com.nearby.chat.feature.account

data class EditMobileUiState(
    val step: Int = STEP_VERIFY_OLD,
    val currentMobile: String = "",
    val ticket: String = "",
    val oldCode: String = "",
    val newMobile: String = "",
    val newCode: String = "",
    val oldCounting: Int = 0,
    val newCounting: Int = 0,
    val sendingOld: Boolean = false,
    val sendingNew: Boolean = false,
    val verifying: Boolean = false,
    val submitting: Boolean = false,
) {
    companion object {
        const val STEP_VERIFY_OLD = 1
        const val STEP_BIND_NEW = 2
    }
}
