package com.nearby.chat.core.media

import android.content.Context
import android.media.MediaRecorder
import android.os.Build
import java.io.File

class VoiceRecorder(
    private val context: Context,
    private val minDurationMs: Long,
    private val maxDurationMs: Long,
) {
    private var recorder: MediaRecorder? = null
    private var file: File? = null
    private var startedAt: Long = 0L

    fun start(): Boolean {
        if (recorder != null) return false
        val out = File(context.cacheDir, "voice_${System.currentTimeMillis()}.m4a")
        val mr = runCatching {
            @Suppress("DEPRECATION")
            val r = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                MediaRecorder(context)
            } else {
                MediaRecorder()
            }
            r.setAudioSource(MediaRecorder.AudioSource.MIC)
            r.setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
            r.setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
            r.setAudioSamplingRate(SAMPLE_RATE)
            r.setAudioEncodingBitRate(BIT_RATE)
            r.setOutputFile(out.absolutePath)
            r.prepare()
            r.start()
            r
        }.getOrElse {
            out.delete()
            return false
        }
        file = out
        recorder = mr
        startedAt = System.currentTimeMillis()
        return true
    }

    fun finish(): VoiceRecording {
        val out = file
        val begin = startedAt
        release()
        if (out == null || !out.exists() || out.length() <= 0L) {
            out?.delete()
            return VoiceRecording.Failed
        }
        val durationMs = (System.currentTimeMillis() - begin).coerceAtLeast(0L)
        if (durationMs < minDurationMs) {
            out.delete()
            return VoiceRecording.TooShort
        }
        return VoiceRecording.Done(out, durationMs.coerceAtMost(maxDurationMs))
    }

    fun cancel() {
        val out = file
        release()
        out?.delete()
    }

    private fun release() {
        val r = recorder
        recorder = null
        file = null
        startedAt = 0L
        if (r != null) {
            runCatching { r.stop() }
            runCatching { r.reset() }
            runCatching { r.release() }
        }
    }

    private companion object {
        const val SAMPLE_RATE = 44_100
        const val BIT_RATE = 96_000
    }
}
