package com.nearby.chat.core.media

import java.io.File

sealed interface VoiceRecording {
    data class Done(val file: File, val durationMs: Long) : VoiceRecording
    data object Failed : VoiceRecording
    data object TooShort : VoiceRecording
}
