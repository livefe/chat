package com.nearby.chat.feature.chat.ui

import android.content.Context
import coil.decode.VideoFrameDecoder
import coil.request.ImageRequest
import coil.request.videoFrameMillis

private const val VIDEO_COVER_FRAME_MS = 400L

internal fun videoCoverRequest(context: Context, url: String): ImageRequest = ImageRequest.Builder(context)
    .data(url)
    .allowHardware(false)
    .videoFrameMillis(VIDEO_COVER_FRAME_MS)
    .decoderFactory(VideoFrameDecoder.Factory())
    .crossfade(true)
    .build()
