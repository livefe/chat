package com.nearby.chat.feature.chat.ui

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.platform.LocalTextToolbar
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.nearby.chat.R
import com.nearby.chat.core.chat.VOICE_MAX_DURATION_MS
import com.nearby.chat.core.chat.VOICE_MIN_DURATION_MS
import com.nearby.chat.core.media.VoiceRecorder
import com.nearby.chat.core.media.VoiceRecording
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.theme.AppColors
import com.nearby.chat.core.theme.AppSpacing
import com.nearby.chat.core.theme.AppType
import com.nearby.chat.core.ui.ComposerSendButton
import com.nearby.chat.core.ui.LocalToaster
import com.nearby.chat.core.ui.toUiText
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

private val ComposerRowVPad = 10.dp
private val InputMinMaxHeight = 120.dp
private val InputMaxMaxHeight = 280.dp
private const val INPUT_MAX_SCREEN_FRACTION = 0.3f
private const val DISABLED_ICON_ALPHA = 0.35f

@Composable
fun ChatComposerBar(
    input: String,
    onInputChange: (String) -> Unit,
    onSend: () -> Unit,
    onImagePicked: (Uri) -> Unit,
    onPickLocation: () -> Unit,
    onVoiceRecorded: (contentJson: String) -> Unit,
    uploadVoice: suspend (Uri, durationMs: Long) -> String,
    sendEnabled: Boolean,
    moreOpen: Boolean,
    onMoreOpenChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
    canCompose: Boolean = true,
    uploading: Boolean = false,
    placeholder: String = stringResource(R.string.composer_placeholder),
    replyTo: MessageItem? = null,
    onClearReply: () -> Unit = {},
    enableVoice: Boolean = true,
    enableLocation: Boolean = true,
    onFilePicked: ((Uri) -> Unit)? = null,
) {
    val context = LocalContext.current
    val toaster = LocalToaster.current
    val focusManager = LocalFocusManager.current
    val keyboard = LocalSoftwareKeyboardController.current
    val textToolbar = LocalTextToolbar.current
    val scope = rememberCoroutineScope()
    val screenHeightDp = LocalConfiguration.current.screenHeightDp
    val inputMaxHeight = (screenHeightDp * INPUT_MAX_SCREEN_FRACTION).dp.coerceIn(InputMinMaxHeight, InputMaxMaxHeight)
    val recorder = remember {
        VoiceRecorder(context.applicationContext, VOICE_MIN_DURATION_MS, VOICE_MAX_DURATION_MS)
    }

    var voiceMode by remember { mutableStateOf(false) }
    val inVoiceMode = enableVoice && voiceMode
    var recording by remember { mutableStateOf(false) }
    var canceling by remember { mutableStateOf(false) }
    var uploadingVoice by remember { mutableStateOf(false) }
    var maxDurationJob by remember { mutableStateOf<Job?>(null) }

    val pickMedia = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickVisualMedia(),
    ) { uri ->
        if (uri == null) return@rememberLauncherForActivityResult
        val cached = runCatching { cachePickedMedia(context, uri) }.getOrElse {
            toaster.fail((it as? PickedMediaException)?.text ?: it.toUiText(R.string.picker_read_album_failed))
            return@rememberLauncherForActivityResult
        }
        onImagePicked(cached)
    }

    val pickFile = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.OpenDocument(),
    ) { uri ->
        if (uri == null || onFilePicked == null) return@rememberLauncherForActivityResult
        val cached = runCatching { cachePickedMedia(context, uri) }.getOrElse {
            toaster.fail((it as? PickedMediaException)?.text ?: it.toUiText(R.string.picker_read_file_failed))
            return@rememberLauncherForActivityResult
        }
        onFilePicked(cached)
    }

    val recordPermission = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission(),
    ) { granted ->
        if (!granted) toaster.fail(R.string.voice_permission_required)
    }

    fun closeMore() {
        onMoreOpenChange(false)
    }

    fun hideInputChrome() {
        textToolbar.hide()
        focusManager.clearFocus(force = true)
        keyboard?.hide()
    }

    fun openMore() {
        if (!canCompose) return
        hideInputChrome()
        onMoreOpenChange(true)
    }

    fun resetRecordingState() {
        maxDurationJob?.cancel()
        maxDurationJob = null
        recording = false
        canceling = false
    }

    fun cancelRecording(silent: Boolean = false) {
        resetRecordingState()
        recorder.cancel()
        if (!silent) toaster.show(R.string.voice_cancelled)
    }

    fun finishRecording() {
        resetRecordingState()
        when (val result = recorder.finish()) {
            VoiceRecording.Failed -> toaster.fail(R.string.voice_record_failed)

            VoiceRecording.TooShort -> toaster.show(R.string.voice_too_short)

            is VoiceRecording.Done -> {
                val uri = Uri.fromFile(result.file)
                uploadingVoice = true
                scope.launch {
                    runCatching { uploadVoice(uri, result.durationMs) }
                        .onSuccess(onVoiceRecorded)
                        .onFailure { e -> toaster.fail(e, R.string.voice_upload_failed) }
                    result.file.delete()
                    uploadingVoice = false
                }
            }
        }
    }

    fun startRecording() {
        if (!canCompose || uploading || uploadingVoice || recording || !inVoiceMode) return
        closeMore()
        hideInputChrome()
        if (!recorder.start()) {
            toaster.fail(R.string.voice_record_failed)
            return
        }
        recording = true
        canceling = false
        maxDurationJob = scope.launch {
            delay(VOICE_MAX_DURATION_MS)
            if (recording) finishRecording()
        }
    }

    DisposableEffect(Unit) {
        onDispose { cancelRecording(silent = true) }
    }

    if (moreOpen) {
        BackHandler { closeMore() }
    }

    LaunchedEffect(canCompose) {
        if (!canCompose) {
            if (moreOpen) closeMore()
            if (recording) cancelRecording(silent = true)
            if (voiceMode) voiceMode = false
        }
    }

    LaunchedEffect(voiceMode) {
        if (voiceMode) {
            closeMore()
            hideInputChrome()
        } else if (recording) {
            cancelRecording(silent = true)
        }
    }

    val busy = uploading || uploadingVoice || recording
    val iconTint = if (!canCompose || busy) {
        AppColors.ComposerIcon.copy(alpha = DISABLED_ICON_ALPHA)
    } else {
        AppColors.ComposerIcon
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(AppColors.PageBg),
    ) {
        if (replyTo != null) {
            ComposerReplyBar(replyTo = replyTo, onClear = onClearReply)
        }

        if (recording) {
            Text(
                text = stringResource(if (canceling) R.string.voice_release_to_cancel else R.string.voice_release_to_send),
                fontSize = AppType.Caption,
                color = if (canceling) ComposerDanger else AppColors.TextMuted,
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = AppSpacing.M, vertical = AppSpacing.Xs),
            )
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = AppSpacing.Xs, vertical = ComposerRowVPad),
            verticalAlignment = Alignment.Bottom,
            horizontalArrangement = Arrangement.spacedBy(AppSpacing.Xs),
        ) {
            if (enableVoice) {
                ComposerVoiceToggle(
                    inVoiceMode = inVoiceMode,
                    enabled = canCompose && !busy,
                    tint = iconTint,
                    onToggle = { voiceMode = !voiceMode },
                )
            }

            if (inVoiceMode) {
                ComposerHoldToTalk(
                    enabled = canCompose && !uploading && !uploadingVoice,
                    recording = recording,
                    canceling = canceling,
                    isRecordingNow = { recording },
                    hasPermission = {
                        context.checkSelfPermission(Manifest.permission.RECORD_AUDIO) ==
                            PackageManager.PERMISSION_GRANTED
                    },
                    onRequestPermission = { recordPermission.launch(Manifest.permission.RECORD_AUDIO) },
                    onStart = ::startRecording,
                    onCancelingChange = { canceling = it },
                    onRelease = { cancel -> if (cancel) cancelRecording() else finishRecording() },
                    modifier = Modifier.weight(1f),
                )
            } else {
                ComposerTextInput(
                    input = input,
                    placeholder = placeholder,
                    maxHeight = inputMaxHeight,
                    onValueChange = {
                        if (moreOpen) closeMore()
                        textToolbar.hide()
                        onInputChange(it)
                    },
                    onFocused = ::closeMore,
                    modifier = Modifier.weight(1f),
                )
            }

            val showSend = !inVoiceMode && input.isNotBlank()
            if (showSend) {
                ComposerSendButton(
                    onClick = {
                        closeMore()
                        textToolbar.hide()
                        onSend()
                    },
                    enabled = sendEnabled,
                )
            } else {
                ComposerMoreToggle(
                    moreOpen = moreOpen,
                    enabled = canCompose && !busy,
                    tint = iconTint,
                    onClick = { if (moreOpen) closeMore() else openMore() },
                )
            }
        }

        if (moreOpen) {
            ComposerMorePanel(
                enabled = canCompose && !busy,
                canCompose = canCompose,
                showLocation = enableLocation,
                showFile = onFilePicked != null,
                onPickMedia = {
                    pickMedia.launch(PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageAndVideo))
                    closeMore()
                },
                onPickLocation = {
                    onPickLocation()
                    closeMore()
                },
                onPickFile = {
                    pickFile.launch(arrayOf("*/*"))
                    closeMore()
                },
            )
        }
    }
}
