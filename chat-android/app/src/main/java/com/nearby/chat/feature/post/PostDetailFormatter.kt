package com.nearby.chat.feature.post

import android.content.Context
import com.nearby.chat.R
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.format.MoneyFormat
import com.nearby.chat.core.format.formatUnixTime
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.model.PostExtras
import com.nearby.chat.core.model.PostItem
import com.nearby.chat.core.post.PostCategories
import com.nearby.chat.core.ui.UiText

object PostDetailFormatter {
    const val FEE_TYPE_CHARGE = 1
    const val FEE_TYPE_REWARD = 2
    private const val CHAT_PEEK_MAX = 28
    private val WHITESPACE = Regex("\\s+")

    fun chatPeek(context: Context, messages: List<MessageItem>): String {
        val last = messages.lastOrNull { it.id > 0 && !it.recalled && it.content.isNotBlank() } ?: return ""
        return chatPeek(context, last.fromName, last.contentType, last.content)
    }

    fun chatPeek(context: Context, fromName: String, contentType: Int, content: String): String {
        if (content.isBlank()) return ""
        val body = messagePreview(contentType, content).resolve(context).replace(WHITESPACE, " ").trim()
        if (body.isEmpty()) return ""
        val preview = if (body.length > CHAT_PEEK_MAX) body.take(CHAT_PEEK_MAX) + "…" else body
        return if (fromName.isBlank()) preview else "$fromName: $preview"
    }

    fun idlePrice(post: PostItem): UiText? {
        if (post.category != PostCategories.IDLE) return null
        val extras = post.extras
        val amount = extras?.amount ?: 0.0
        val negotiable = extras?.negotiable == true
        return when {
            amount > 0 && negotiable -> UiText.of(R.string.price_amount_negotiable_detail, MoneyFormat.formatAmount(amount))
            amount > 0 -> UiText.of(R.string.price_amount, MoneyFormat.formatAmount(amount))
            negotiable -> UiText.of(R.string.price_negotiable)
            else -> null
        }
    }

    fun hasFee(extras: PostExtras?): Boolean = extras != null && extras.feeType in FEE_TYPE_CHARGE..FEE_TYPE_REWARD && extras.amount > 0

    fun hasExtras(post: PostItem): Boolean {
        if (post.startAt > 0 || post.endAt > 0) return true
        if (post.extras?.needSignup == true) return true
        if (hasFee(post.extras)) return true
        return idlePrice(post) != null
    }

    fun quota(extras: PostExtras): UiText = if (extras.allowWaitlist) {
        UiText.of(R.string.quota_limit_waitlist, extras.maxParticipants)
    } else {
        UiText.of(R.string.quota_limit, extras.maxParticipants)
    }

    fun signupSummary(post: PostItem, max: Int): UiText {
        val confirmed = post.signupSummary?.confirmed ?: 0
        val waitlisted = post.signupSummary?.waitlisted ?: 0
        if (max > 0) {
            return if (waitlisted > 0) {
                UiText.of(R.string.signup_count_limit_waitlist, confirmed, max, waitlisted)
            } else {
                UiText.of(R.string.signup_count_limit, confirmed, max)
            }
        }
        return when {
            confirmed > 0 && waitlisted > 0 -> UiText.of(R.string.signup_count_both, confirmed, waitlisted)
            confirmed > 0 -> UiText.of(R.string.signup_count_confirmed, confirmed)
            waitlisted > 0 -> UiText.of(R.string.signup_count_waitlisted, waitlisted)
            else -> UiText.of(R.string.signup_count_none)
        }
    }

    fun fee(extras: PostExtras): UiText {
        val amount = MoneyFormat.formatAmount(extras.amount)
        return if (extras.feeType == FEE_TYPE_REWARD) {
            UiText.of(R.string.fee_reward_amount, amount)
        } else {
            UiText.of(R.string.fee_charge_amount, amount)
        }
    }

    fun createdAt(raw: String): String {
        if (raw.isBlank()) return ""
        return if (raw.length >= 16) raw.substring(5, 16) else raw
    }

    fun unixTime(context: Context, sec: Long): String = formatUnixTime(context, sec, R.string.time_pattern_month_day_hm)
}
