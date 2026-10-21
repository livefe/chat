package com.nearby.chat.core.ui

import android.content.ClipboardManager
import android.content.Context
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.Box
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.Stable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalTextToolbar
import androidx.compose.ui.platform.TextToolbar
import androidx.compose.ui.platform.TextToolbarStatus
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.AnnotatedString
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.nearby.chat.R
import kotlin.math.abs
import androidx.compose.ui.platform.ClipboardManager as ComposeClipboard

interface TextFieldHoldMenu {
    fun beginFingerHold()
    fun endFingerHold()
}

val LocalTextFieldHoldMenu = staticCompositionLocalOf<TextFieldHoldMenu?> { null }

@Stable
private class AppTextToolbar(
    private val model: EditToolbarModel,
) : TextToolbar,
    TextFieldHoldMenu {
    private var suppressHide = false
    private var shownDuringHold = false

    override fun beginFingerHold() {
        suppressHide = true
        shownDuringHold = false
        model.allowOutsideDismiss = false
    }

    override fun endFingerHold() {
        val keep = shownDuringHold
        suppressHide = false
        shownDuringHold = false
        model.allowOutsideDismiss = true
        if (!keep) {
            status = TextToolbarStatus.Hidden
            model.hideFromToolbar()
        }
    }
    override var status: TextToolbarStatus = TextToolbarStatus.Hidden
        private set

    override fun showMenu(
        rect: Rect,
        onCopyRequested: (() -> Unit)?,
        onPasteRequested: (() -> Unit)?,
        onCutRequested: (() -> Unit)?,
        onSelectAllRequested: (() -> Unit)?,
    ) {
        if (model.consumeIgnoreShow()) return
        shownDuringHold = true
        status = TextToolbarStatus.Shown
        model.show(rect, onCopyRequested, onPasteRequested, onCutRequested, onSelectAllRequested)
    }

    override fun hide() {
        if (suppressHide) return
        status = TextToolbarStatus.Hidden
        model.hideFromToolbar()
    }

    fun dismissByUser() {
        if (suppressHide) return
        status = TextToolbarStatus.Hidden
        model.hideFromUser()
    }
}

@Stable
private class EditToolbarModel {
    var onCopy: (() -> Unit)? = null
    var onPaste: (() -> Unit)? = null
    var onCut: (() -> Unit)? = null
    var onSelectAll: (() -> Unit)? = null
    var visible by mutableStateOf(false)
    var rect by mutableStateOf(Rect.Zero)
    var canPaste by mutableStateOf(false)
    var canCopy by mutableStateOf(false)
    var canCut by mutableStateOf(false)
    var canSelectAll by mutableStateOf(false)
    var allowOutsideDismiss by mutableStateOf(true)
    var ignoreShow: Boolean = false

    fun consumeIgnoreShow(): Boolean {
        if (!ignoreShow) return false
        ignoreShow = false
        return true
    }

    fun show(
        newRect: Rect,
        copy: (() -> Unit)?,
        paste: (() -> Unit)?,
        cut: (() -> Unit)?,
        selectAll: (() -> Unit)?,
    ) {
        onCopy = copy
        onPaste = paste
        onCut = cut
        onSelectAll = selectAll
        val nextPaste = paste != null
        val nextCopy = copy != null
        val nextCut = cut != null
        val nextSelectAll = selectAll != null
        if (canPaste != nextPaste) canPaste = nextPaste
        if (canCopy != nextCopy) canCopy = nextCopy
        if (canCut != nextCut) canCut = nextCut
        if (canSelectAll != nextSelectAll) canSelectAll = nextSelectAll
        if (!rectNearlySame(rect, newRect)) rect = newRect
        if (!visible) visible = true
    }

    fun hideFromToolbar() {
        hideInternal()
    }

    fun hideFromUser() {
        ignoreShow = true
        hideInternal()
    }

    private fun hideInternal() {
        if (!visible) return
        visible = false
        onCopy = null
        onPaste = null
        onCut = null
        onSelectAll = null
    }
}

private fun rectNearlySame(a: Rect, b: Rect): Boolean = abs(a.left - b.left) < 0.5f &&
    abs(a.top - b.top) < 0.5f &&
    abs(a.right - b.right) < 0.5f &&
    abs(a.bottom - b.bottom) < 0.5f

private class TextClipboard(
    private val context: Context,
    private val delegate: ComposeClipboard,
) : ComposeClipboard {
    private val system = context.getSystemService(ClipboardManager::class.java)

    override fun getText(): AnnotatedString? {
        val clip = system?.primaryClip
        if (clip != null && clip.itemCount > 0) {
            val text = clip.getItemAt(0).coerceToText(context)?.toString().orEmpty()
            if (text.isNotEmpty()) return AnnotatedString(text)
        }
        return delegate.getText()?.takeIf { it.text.isNotEmpty() }
    }

    override fun setText(annotatedString: AnnotatedString) {
        delegate.setText(annotatedString)
    }

    override fun hasText(): Boolean {
        val description = system?.primaryClipDescription
        if (description?.hasMimeType("text/*") == true) return true
        return getText()?.text?.isNotEmpty() == true
    }
}

@Composable
fun ProvideAppTextToolbar(content: @Composable () -> Unit) {
    val model = remember { EditToolbarModel() }
    val textToolbar = remember(model) { AppTextToolbar(model) }
    val context = LocalContext.current
    val platformClipboard = LocalClipboardManager.current
    val clipboard = remember(context, platformClipboard) { TextClipboard(context, platformClipboard) }
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner, textToolbar) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_PAUSE) textToolbar.hide()
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            textToolbar.hide()
        }
    }
    CompositionLocalProvider(
        LocalTextToolbar provides textToolbar,
        LocalTextFieldHoldMenu provides textToolbar,
        LocalClipboardManager provides clipboard,
    ) {
        Box {
            content()
            AppEditToolbarPopup(model = model, onHide = { textToolbar.dismissByUser() })
        }
    }
}

@Composable
private fun AppEditToolbarPopup(
    model: EditToolbarModel,
    onHide: () -> Unit,
) {
    if (!model.visible) return
    BackHandler { onHide() }
    val pasteLabel = stringResource(R.string.text_action_paste)
    val copyLabel = stringResource(R.string.common_copy)
    val cutLabel = stringResource(R.string.text_action_cut)
    val selectAllLabel = stringResource(R.string.text_action_select_all)
    val actions = buildList {
        if (model.canPaste) {
            add(
                DarkActionItem(pasteLabel) {
                    model.onPaste?.invoke()
                    onHide()
                },
            )
        }
        if (model.canCopy) {
            add(
                DarkActionItem(copyLabel) {
                    model.onCopy?.invoke()
                    onHide()
                },
            )
        }
        if (model.canCut) {
            add(
                DarkActionItem(cutLabel) {
                    model.onCut?.invoke()
                    onHide()
                },
            )
        }
        if (model.canSelectAll) {
            add(
                DarkActionItem(selectAllLabel) {
                    model.onSelectAll?.invoke()
                    onHide()
                },
            )
        }
    }
    DarkActionPillPopup(
        anchor = model.rect,
        actions = actions,
        onDismiss = onHide,
        focusable = false,
        dismissOnClickOutside = model.allowOutsideDismiss,
    )
}
