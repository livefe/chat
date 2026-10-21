package com.nearby.chat.core.post

import androidx.annotation.StringRes
import com.nearby.chat.R
import com.nearby.chat.core.model.PostItem

object PostSignupRules {
    const val DEADLINE_GRACE_SECONDS = 59L
    const val STATUS_CONFIRMED = 1
    const val STATUS_WAITLISTED = 2
    const val STATUS_CANCELLED = 3

    fun isSignupDeadlinePassed(deadlineSec: Long, nowSec: Long = System.currentTimeMillis() / 1000): Boolean {
        if (deadlineSec <= 0) return false
        return nowSec > deadlineSec + DEADLINE_GRACE_SECONDS
    }

    fun isEventEnded(endAt: Long, nowSec: Long = System.currentTimeMillis() / 1000): Boolean = endAt > 0 && nowSec > endAt

    fun isCheckInWindowOpen(
        startAt: Long,
        endAt: Long,
        nowSec: Long = System.currentTimeMillis() / 1000,
    ): Boolean {
        if (startAt > 0 && nowSec < startAt) return false
        if (endAt > 0 && nowSec > endAt) return false
        return true
    }

    fun schedulePhase(
        startAt: Long,
        endAt: Long,
        nowSec: Long = System.currentTimeMillis() / 1000,
    ): SchedulePhase {
        if (startAt <= 0 && endAt <= 0) return SchedulePhase.None
        if (endAt > 0 && nowSec > endAt) return SchedulePhase.Ended
        if (startAt > 0 && nowSec < startAt) return SchedulePhase.NotStarted
        if (startAt > 0 && (endAt <= 0 || nowSec <= endAt)) return SchedulePhase.Ongoing
        if (endAt > 0 && nowSec <= endAt) return SchedulePhase.Ongoing
        return SchedulePhase.None
    }

    fun hasTimestamp(raw: String): Boolean = raw.isNotBlank()

    fun hasTimestamp(ts: Long): Boolean = ts > 0

    @StringRes
    fun checkInLabel(checkedIn: Boolean): Int = if (checkedIn) R.string.signup_checked_in else R.string.signup_not_checked_in

    fun canShowJoinedCheckInCode(post: PostItem): Boolean {
        if (post.status != 1) return false
        val extras = post.extras ?: return false
        if (!extras.needSignup || !extras.needCheckIn) return false
        if (post.mySignupStatus != STATUS_CONFIRMED) return false
        if (hasTimestamp(post.myCheckedInAt)) return true
        return isCheckInWindowOpen(post.startAt, post.endAt)
    }

    fun canShowJoinedCheckInCodeIcon(post: PostItem): Boolean = canShowJoinedCheckInCode(post) && !hasTimestamp(post.myCheckedInAt)

    fun canDiscuss(post: PostItem): Boolean = post.status == PostStatus.PUBLISHED && !isEventEnded(post.endAt)

    fun canDeletePublishedPost(post: PostItem): Boolean = post.status == 0 || post.status == 2 || post.status == 3

    fun canDeleteJoinedPost(post: PostItem): Boolean {
        if (post.status == 3) return true
        if (hasTimestamp(post.myCheckedInAt)) return true
        return isEventEnded(post.endAt)
    }

    const val CHECK_IN_NOT_CHECKED = 1
    const val CHECK_IN_CHECKED = 2
}
