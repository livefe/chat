package com.nearby.chat.feature.report

import androidx.annotation.StringRes
import com.nearby.chat.R

enum class ReportReason(val value: Int, @StringRes val label: Int) {
    SPAM(1, R.string.report_reason_spam),
    PORN(2, R.string.report_reason_porn),
    ABUSE(3, R.string.report_reason_abuse),
    ILLEGAL(4, R.string.report_reason_illegal),
    FAKE(5, R.string.report_reason_fake),
    INFRINGE(6, R.string.report_reason_infringe),
    OTHER(7, R.string.report_reason_other),
}
