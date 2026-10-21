package com.nearby.chat.core.media

import android.media.MediaPlayer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class VoicePlayer @Inject constructor() {
    private var player: MediaPlayer? = null
    private var currentKey: String = ""

    private val _playingKey = MutableStateFlow("")
    val playingKey: StateFlow<String> = _playingKey.asStateFlow()

    fun isPlaying(key: String): Boolean {
        if (key.isBlank() || currentKey != key) return false
        val p = player ?: return false
        return runCatching { p.isPlaying }.getOrDefault(false)
    }

    fun stop() {
        val p = player
        player = null
        currentKey = ""
        if (p != null) {
            runCatching {
                if (p.isPlaying) p.stop()
            }
            runCatching { p.reset() }
            runCatching { p.release() }
        }
        publish()
    }

    fun toggle(key: String, url: String) {
        if (key.isBlank() || url.isBlank()) return
        if (currentKey == key && isPlaying(key)) {
            stop()
            return
        }
        stop()
        currentKey = key
        publish()
        val mp = MediaPlayer()
        player = mp
        mp.setOnCompletionListener { stop() }
        mp.setOnErrorListener { _, _, _ ->
            stop()
            true
        }
        runCatching {
            mp.setDataSource(url)
            mp.prepareAsync()
            mp.setOnPreparedListener {
                if (player !== mp || currentKey != key) {
                    runCatching { mp.release() }
                    return@setOnPreparedListener
                }
                runCatching { mp.start() }
                    .onFailure { stop() }
                    .onSuccess { publish() }
            }
        }.onFailure {
            stop()
        }
    }

    private fun publish() {
        val key = currentKey
        _playingKey.value = if (isPlaying(key)) key else ""
    }
}
