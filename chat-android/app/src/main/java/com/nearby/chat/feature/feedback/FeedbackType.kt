package com.nearby.chat.feature.feedback

import androidx.annotation.StringRes
import com.nearby.chat.R

enum class FeedbackType(val value: Int, @StringRes val label: Int) {
    SUGGESTION(1, R.string.feedback_type_suggestion),
    BUG(2, R.string.feedback_type_bug),
    EXPERIENCE(3, R.string.feedback_type_experience),
    ACCOUNT(4, R.string.feedback_type_account),
    OTHER(5, R.string.feedback_type_other),
}
