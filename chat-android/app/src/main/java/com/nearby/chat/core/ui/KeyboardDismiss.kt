package com.nearby.chat.core.ui

import android.view.View
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.Modifier
import androidx.compose.ui.composed
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.input.pointer.PointerEventPass
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.LayoutCoordinates
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.platform.SoftwareKeyboardController
import androidx.compose.ui.semantics.SemanticsActions
import androidx.compose.ui.semantics.SemanticsNode
import androidx.compose.ui.semantics.SemanticsOwner
import androidx.compose.ui.semantics.getOrNull
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull

@Composable
fun Modifier.dismissKeyboardOnTap(onTap: (() -> Unit)? = null): Modifier {
    val focusManager = LocalFocusManager.current
    val keyboard = LocalSoftwareKeyboardController.current
    val tap = rememberUpdatedState(onTap)
    val tracker = LocalTextInputPressTracker.current
    return pointerInput(Unit) {
        awaitEachGesture {
            val down = awaitFirstDown(requireUnconsumed = false)
            val slop = viewConfiguration.touchSlop
            val start = down.position
            var moved = false
            while (true) {
                val event = awaitPointerEvent(PointerEventPass.Final)
                val change = event.changes.firstOrNull { it.id == down.id } ?: break
                if ((change.position - start).getDistance() > slop) moved = true
                if (!change.pressed) {
                    if (!moved) {
                        tap.value?.invoke()
                        if (tracker?.pressed != true) {
                            focusManager.clearFocus(force = true)
                            keyboard?.hide()
                        }
                    }
                    break
                }
            }
        }
    }
}

private class FocusFlag {
    var value = false
}

private class CoordFlag {
    var value: LayoutCoordinates? = null
}

private class HoldSession {
    var active = true
}

private class ImeHideGate {
    private var token = 0

    fun next(): Int {
        token += 1
        return token
    }

    fun isCurrent(value: Int) = token == value
}

private enum class FieldPress {
    Tap,
    Long,
    Moved,
}

fun Modifier.imeOnTap(): Modifier = composed {
    val view = LocalView.current
    val keyboardState = rememberUpdatedState(LocalSoftwareKeyboardController.current)
    val holdMenuState = rememberUpdatedState(LocalTextFieldHoldMenu.current)
    val gate = remember { ImeHideGate() }
    val focused = remember { FocusFlag() }
    val coordinates = remember { CoordFlag() }
    onGloballyPositioned { coordinates.value = it }
        .onFocusChanged { focused.value = it.isFocused }
        .pointerInput(Unit) {
            coroutineScope {
                val scope = this
                awaitEachGesture {
                    val down = awaitFirstDown(
                        requireUnconsumed = false,
                        pass = PointerEventPass.Initial,
                    )
                    val wasFocused = focused.value
                    val slop = viewConfiguration.touchSlop
                    val start = down.position
                    val press = withTimeoutOrNull(viewConfiguration.longPressTimeoutMillis) {
                        while (true) {
                            val event = awaitPointerEvent(PointerEventPass.Initial)
                            val change = event.changes.firstOrNull { it.id == down.id }
                                ?: return@withTimeoutOrNull FieldPress.Moved
                            val dx = change.position.x - start.x
                            val dy = change.position.y - start.y
                            if (dx * dx + dy * dy > slop * slop) return@withTimeoutOrNull FieldPress.Moved
                            if (!change.pressed) return@withTimeoutOrNull FieldPress.Tap
                        }
                    } ?: FieldPress.Long
                    val keyboard = keyboardState.value
                    when (press) {
                        FieldPress.Tap -> {
                            gate.next()
                            keyboard?.show()
                        }

                        FieldPress.Long -> {
                            if (!wasFocused) scope.suppressIme(gate, keyboard)
                            val menu = holdMenuState.value
                            val session = HoldSession()
                            val placed = coordinates.value
                            val rootPos = if (placed != null && placed.isAttached) placed.localToRoot(start) else null
                            menu?.beginFingerHold()
                            val showJob = scope.launch {
                                withFrameNanos { }
                                withFrameNanos { }
                                if (!session.active || rootPos == null) return@launch
                                view.performTextFieldLongClick(rootPos)
                            }
                            try {
                                while (true) {
                                    val event = awaitPointerEvent(PointerEventPass.Initial)
                                    val change = event.changes.firstOrNull { it.id == down.id } ?: break
                                    if (!change.pressed) break
                                }
                            } finally {
                                session.active = false
                                showJob.cancel()
                                menu?.endFingerHold()
                            }
                        }

                        FieldPress.Moved -> Unit
                    }
                }
            }
        }
}

private fun View.performTextFieldLongClick(rootPos: Offset) {
    val owner = composeSemanticsOwner() ?: return
    val node = owner.unmergedRootSemanticsNode.findLongClick(rootPos) ?: return
    node.config.getOrNull(SemanticsActions.OnLongClick)?.action?.invoke()
}

private fun View.composeSemanticsOwner(): SemanticsOwner? {
    val method = javaClass.methods.firstOrNull { it.name == "getSemanticsOwner" && it.parameterCount == 0 }
        ?: return null
    return method.invoke(this) as? SemanticsOwner
}

private fun SemanticsNode.findLongClick(rootPos: Offset): SemanticsNode? {
    if (!boundsInRoot.contains(rootPos)) return null
    for (child in children) {
        child.findLongClick(rootPos)?.let { return it }
    }
    if (config.getOrNull(SemanticsActions.OnLongClick)?.action == null) return null
    return this
}

private fun CoroutineScope.suppressIme(
    gate: ImeHideGate,
    keyboard: SoftwareKeyboardController?,
) {
    val token = gate.next()
    launch {
        repeat(4) {
            if (!gate.isCurrent(token)) return@launch
            keyboard?.hide()
            withFrameNanos { }
        }
    }
}

fun Modifier.markTextInputPress(): Modifier = composed {
    val tracker = LocalTextInputPressTracker.current ?: return@composed this
    pointerInput(tracker) {
        awaitEachGesture {
            awaitFirstDown(requireUnconsumed = false)
            tracker.pressed = true
            do {
                val event = awaitPointerEvent(PointerEventPass.Final)
            } while (event.changes.any { it.pressed })
            tracker.pressed = false
        }
    }
}
