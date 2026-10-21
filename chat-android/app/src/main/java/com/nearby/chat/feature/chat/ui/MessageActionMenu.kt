package com.nearby.chat.feature.chat.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.nearby.chat.R
import com.nearby.chat.core.ui.DarkActionItem
import com.nearby.chat.core.ui.DarkActionPill

@Composable
fun MessageActionMenu(
    onCopy: () -> Unit,
    onQuote: () -> Unit,
    onForward: (() -> Unit)? = null,
    onRecall: (() -> Unit)? = null,
    onReport: (() -> Unit)? = null,
) {
    val copy = stringResource(R.string.common_copy)
    val forward = stringResource(R.string.forward_title)
    val quote = stringResource(R.string.message_action_quote)
    val recall = stringResource(R.string.message_action_recall)
    val report = stringResource(R.string.message_action_report)
    DarkActionPill(
        buildList {
            add(DarkActionItem(copy, onCopy))
            if (onForward != null) add(DarkActionItem(forward, onForward))
            add(DarkActionItem(quote, onQuote))
            if (onRecall != null) add(DarkActionItem(recall, onRecall))
            if (onReport != null) add(DarkActionItem(report, onReport))
        },
    )
}
