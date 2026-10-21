package com.nearby.chat.core.post

import com.nearby.chat.R
import com.nearby.chat.core.format.MoneyFormat
import com.nearby.chat.core.model.PostExtras
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.model.PostSignupSummary
import com.nearby.chat.core.ui.UiText
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale

object PostListHighlights {
    fun groups(post: PostItem): PostListFactGroups {
        val extras = post.extras
        val nowSec = System.currentTimeMillis() / 1000

        val phase = when (PostSignupRules.schedulePhase(post.startAt, post.endAt, nowSec)) {
            SchedulePhase.NotStarted -> PostListTag(UiText.of(R.string.phase_not_started), PostListTagTone.PhasePending)
            SchedulePhase.Ongoing -> PostListTag(UiText.of(R.string.phase_ongoing), PostListTagTone.PhaseLive)
            SchedulePhase.Ended -> PostListTag(UiText.of(R.string.phase_ended), PostListTagTone.PhaseDone)
            SchedulePhase.None -> null
        }

        val schedule = formatScheduleCompact(post.startAt, post.endAt)

        val signup = mutableListOf<UiText>()
        var signupDeadline: UiText? = null
        if (extras?.needSignup == true) {
            signup += signupBrief(extras, post.signupSummary)
            val deadline = extras.signupDeadlineAt
            if (deadline > 0) {
                val text = if (PostSignupRules.isSignupDeadlinePassed(deadline, nowSec)) {
                    UiText.of(R.string.signup_deadline_passed)
                } else {
                    UiText.of(R.string.signup_deadline_at, formatUnixCompact(deadline))
                }
                signupDeadline = text
                signup += text
            }
        }

        val fee = if (post.category == PostCategories.IDLE) {
            idlePriceBrief(extras) ?: feeBrief(extras)
        } else {
            feeBrief(extras)
        }

        return PostListFactGroups(
            phase = phase,
            schedule = schedule,
            signup = signup,
            signupDeadline = signupDeadline,
            fee = fee,
        )
    }

    fun formatScheduleCompact(startAt: Long, endAt: Long): String? {
        if (startAt <= 0 && endAt <= 0) return null
        if (startAt > 0 && endAt > 0) {
            val a = Calendar.getInstance().apply { timeInMillis = startAt * 1000 }
            val b = Calendar.getInstance().apply { timeInMillis = endAt * 1000 }
            val sameDay =
                a.get(Calendar.YEAR) == b.get(Calendar.YEAR) &&
                    a.get(Calendar.DAY_OF_YEAR) == b.get(Calendar.DAY_OF_YEAR)
            if (sameDay) {
                return "${formatDateCompact(a.time)} ${pad(a.get(Calendar.HOUR_OF_DAY))}:${pad(a.get(Calendar.MINUTE))}–${pad(b.get(Calendar.HOUR_OF_DAY))}:${pad(b.get(Calendar.MINUTE))}"
            }
            return "${formatUnixCompact(startAt)}–${formatUnixCompact(endAt)}"
        }
        return formatUnixCompact(if (startAt > 0) startAt else endAt)
    }

    private fun idlePriceBrief(extras: PostExtras?): UiText? {
        if (extras == null) return null
        val amount = extras.amount
        if (amount <= 0 && !extras.negotiable) return null
        if (amount <= 0) return UiText.of(R.string.price_negotiable)
        val price = MoneyFormat.formatAmount(amount)
        return if (extras.negotiable) {
            UiText.of(R.string.price_amount_negotiable, price)
        } else {
            UiText.of(R.string.price_amount, price)
        }
    }

    private fun feeBrief(extras: PostExtras?): UiText? {
        if (extras == null) return null
        if (extras.feeType !in 1..2 || extras.amount <= 0) return null
        val price = MoneyFormat.formatAmount(extras.amount)
        return when (extras.feeType) {
            2 -> UiText.of(R.string.fee_reward_amount, price)
            else -> UiText.of(R.string.fee_charge_amount, price)
        }
    }

    private fun signupBrief(extras: PostExtras, summary: PostSignupSummary?): UiText {
        val confirmed = summary?.confirmed ?: 0
        val waitlisted = summary?.waitlisted ?: 0
        val max = extras.maxParticipants
        return when {
            max > 0 && waitlisted > 0 -> UiText.of(R.string.signup_brief_limit_waitlist, confirmed, max, waitlisted)
            max > 0 -> UiText.of(R.string.signup_brief_limit, confirmed, max)
            confirmed > 0 && waitlisted > 0 -> UiText.of(R.string.signup_brief_count_waitlist, confirmed, waitlisted)
            confirmed > 0 -> UiText.of(R.string.signup_brief_count, confirmed)
            waitlisted > 0 -> UiText.of(R.string.signup_brief_waitlist, waitlisted)
            else -> UiText.of(R.string.signup_required)
        }
    }

    private fun formatUnixCompact(sec: Long): String {
        if (sec <= 0) return ""
        val cal = Calendar.getInstance().apply { timeInMillis = sec * 1000 }
        return "${formatDateCompact(cal.time)} ${pad(cal.get(Calendar.HOUR_OF_DAY))}:${pad(cal.get(Calendar.MINUTE))}"
    }

    private fun formatDateCompact(date: Date): String {
        val thisYear = Calendar.getInstance().get(Calendar.YEAR)
        val cal = Calendar.getInstance().apply { time = date }
        return if (cal.get(Calendar.YEAR) == thisYear) {
            SimpleDateFormat("M-d", Locale.getDefault()).format(date)
        } else {
            SimpleDateFormat("yyyy-M-d", Locale.getDefault()).format(date)
        }
    }

    private fun pad(n: Int): String = n.toString().padStart(2, '0')

    fun formatListCreatedAt(raw: String): String {
        if (raw.isBlank()) return ""
        return if (raw.length >= 16) raw.substring(5, 16) else raw
    }
}
