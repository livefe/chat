package com.nearby.chat.feature.post

import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.model.PostSignupListCounts
import com.nearby.chat.core.post.PostSignupRules

internal data class SignupTab(
    @StringRes val label: Int,
    val status: Int? = null,
    val checkIn: Int? = null,
) {
    companion object {
        fun statusTabs(counts: PostSignupListCounts): List<Pair<SignupTab, Int>> = listOf(
            SignupTab(R.string.common_all) to counts.total,
            SignupTab(R.string.signup_status_confirmed, status = PostSignupRules.STATUS_CONFIRMED) to counts.confirmed,
            SignupTab(R.string.signup_status_waitlisted, status = PostSignupRules.STATUS_WAITLISTED) to counts.waitlisted,
            SignupTab(R.string.signup_status_cancelled, status = PostSignupRules.STATUS_CANCELLED) to counts.cancelled,
        )

        fun checkInTabs(counts: PostSignupListCounts): List<Pair<SignupTab, Int>> = listOf(
            SignupTab(R.string.signup_not_checked_in, checkIn = PostSignupRules.CHECK_IN_NOT_CHECKED) to counts.notCheckedIn,
            SignupTab(R.string.signup_checked_in, checkIn = PostSignupRules.CHECK_IN_CHECKED) to counts.checkedIn,
        )
    }
}
