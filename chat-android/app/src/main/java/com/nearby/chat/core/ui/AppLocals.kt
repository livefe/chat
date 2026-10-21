package com.nearby.chat.core.ui

import androidx.compose.runtime.staticCompositionLocalOf
import com.nearby.chat.core.media.VoicePlayer

val LocalToaster = staticCompositionLocalOf<Toaster> { error("Toaster not provided") }

val LocalVoicePlayer = staticCompositionLocalOf<VoicePlayer> { error("VoicePlayer not provided") }
