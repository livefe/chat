package com.nearby.chat.core.h5

import android.net.Uri
import com.nearby.chat.BuildConfig
import com.nearby.chat.core.share.ShareLinks

object H5Urls {
    fun page(path: String, query: Map<String, String> = emptyMap()): String {
        val base = BuildConfig.H5_BASE_URL.trimEnd('/')
        val normalized = if (path.startsWith("/")) path else "/$path"
        val builder = Uri.parse("$base$normalized").buildUpon()
            .appendQueryParameter(ShareLinks.QUERY_FROM, ShareLinks.FROM_APP)
        query.forEach { (k, v) -> builder.appendQueryParameter(k, v) }
        return builder.build().toString()
    }
}
