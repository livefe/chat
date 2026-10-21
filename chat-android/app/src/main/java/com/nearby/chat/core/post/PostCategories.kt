package com.nearby.chat.core.post

import androidx.annotation.StringRes
import com.nearby.chat.R

object PostCategories {
    const val EVENT = "event"
    const val NOTICE = "notice"
    const val IDLE = "idle"
    const val HELP = "help"

    data class Option(
        val value: String,
        @StringRes val label: Int,
        @StringRes val hint: Int,
    )

    val options = listOf(
        Option(EVENT, R.string.post_category_event, R.string.post_category_event_hint),
        Option(NOTICE, R.string.post_category_notice, R.string.post_category_notice_hint),
        Option(IDLE, R.string.post_category_idle, R.string.post_category_idle_hint),
        Option(HELP, R.string.post_category_help, R.string.post_category_help_hint),
    )

    @StringRes
    fun label(category: String): Int = option(category)?.label ?: R.string.chat_post_default_title

    fun option(category: String): Option? = options.find { it.value == category }
}
