package com.nearby.chat.core.share

import android.content.Intent
import android.net.Uri
import com.nearby.chat.BuildConfig

object ShareLinks {
    const val QUERY_FROM = "from"
    const val QUERY_CONTINUE = "continue"
    const val QUERY_UTM_SOURCE = "utm_source"
    const val QUERY_UTM_MEDIUM = "utm_medium"

    const val FROM_SHARE = "share"
    const val FROM_APP = "app"
    const val UTM_SOURCE = "share"

    const val APP_SCHEME = "nearby"
    const val EXTRA_PEER_ID = "peerId"

    sealed class Target {
        data class Post(val id: Long) : Target()
        data class Room(val id: Long) : Target()
        data class User(val id: Long) : Target()
        data class PrivateChat(val peerId: Long) : Target()
        data class Download(val continuePath: String?) : Target()
    }

    enum class Channel(val value: String) {
        LINK("link"),
        COPY("copy"),
        SYSTEM("system"),
        WECHAT("wechat"),
        XHS("xhs"),
        DOUYIN("douyin"),
        QR("qr"),
        OPEN_APP("open_app"),
        DOWNLOAD("download"),
    }

    fun webPath(target: Target): String = when (target) {
        is Target.Post -> "/post/${target.id}"
        is Target.Room -> "/room/${target.id}"
        is Target.User -> "/user/${target.id}"
        is Target.PrivateChat -> "/messages/${target.peerId}"
        is Target.Download -> "/download"
    }

    fun webUrl(target: Target, channel: Channel = Channel.LINK): String {
        val base = BuildConfig.H5_BASE_URL.trimEnd('/')
        if (base.isBlank()) return ""
        val builder = Uri.parse("$base${webPath(target)}").buildUpon()
        if (target is Target.Download) {
            target.continuePath?.let {
                builder.appendQueryParameter(QUERY_CONTINUE, normalizeAppPath(it))
            }
        }
        builder.appendQueryParameter(QUERY_FROM, FROM_SHARE)
        builder.appendQueryParameter(QUERY_UTM_SOURCE, UTM_SOURCE)
        builder.appendQueryParameter(QUERY_UTM_MEDIUM, channel.value)
        return builder.build().toString()
    }

    fun normalizeAppPath(raw: String): String {
        val path = raw.trim()
        if (!path.startsWith("/")) return "/"
        if (path.startsWith("//") || path.contains("://")) return "/"
        return path.substringBefore('#')
    }

    fun parse(uri: Uri?): Target? {
        if (uri == null) return null
        val scheme = uri.scheme?.lowercase().orEmpty()
        return when (scheme) {
            APP_SCHEME -> parseAppScheme(uri)
            "http", "https" -> parseHttp(uri)
            else -> null
        }
    }

    fun parseAppPath(pathWithQuery: String): Target? {
        val normalized = normalizeAppPath(pathWithQuery)
        val uri = Uri.parse("https://share.local$normalized")
        return parseHttpPath(uri)
    }

    fun parseIntent(intent: Intent?): Target? {
        if (intent == null) return null
        parse(intent.data)?.let { return it }
        val peer = intent.getLongExtra(EXTRA_PEER_ID, 0L)
        return if (peer > 0) Target.PrivateChat(peer) else null
    }

    fun targetKey(target: Target): String = when (target) {
        is Target.Post -> "post:${target.id}"
        is Target.Room -> "room:${target.id}"
        is Target.User -> "user:${target.id}"
        is Target.PrivateChat -> "messages:${target.peerId}"
        is Target.Download -> "download:${target.continuePath.orEmpty()}"
    }

    private fun parseAppScheme(uri: Uri): Target? {
        val host = uri.host.orEmpty().lowercase()
        val id = uri.pathSegments.firstOrNull()?.toLongOrNull()
            ?: uri.lastPathSegment?.toLongOrNull()
        return when (host) {
            "post" -> id?.takeIf { it > 0 }?.let { Target.Post(it) }

            "room" -> id?.takeIf { it > 0 }?.let { Target.Room(it) }

            "user" -> id?.takeIf { it > 0 }?.let { Target.User(it) }

            "messages" -> id?.takeIf { it > 0 }?.let { Target.PrivateChat(it) }

            "download" -> Target.Download(
                uri.getQueryParameter(QUERY_CONTINUE)?.let(::normalizeAppPath),
            )

            else -> null
        }
    }

    private fun parseHttp(uri: Uri): Target? {
        val host = BuildConfig.SHARE_LINK_HOST.trim().lowercase()
        if (host.isNotBlank() && uri.host?.lowercase() != host) return null
        return parseHttpPath(uri)
    }

    private fun parseHttpPath(uri: Uri): Target? {
        val segments = uri.pathSegments.filter { it.isNotBlank() }
        if (segments.isEmpty()) return null
        when (segments[0].lowercase()) {
            "post" -> {
                val id = segments.getOrNull(1)?.toLongOrNull() ?: return null
                if (id <= 0) return null
                return Target.Post(id)
            }

            "room" -> {
                val id = segments.getOrNull(1)?.toLongOrNull() ?: return null
                if (id <= 0) return null
                return Target.Room(id)
            }

            "user" -> {
                val id = segments.getOrNull(1)?.toLongOrNull() ?: return null
                if (id <= 0) return null
                return Target.User(id)
            }

            "messages" -> {
                val id = segments.getOrNull(1)?.toLongOrNull() ?: return null
                if (id <= 0) return null
                return Target.PrivateChat(id)
            }

            "download" -> return Target.Download(
                uri.getQueryParameter(QUERY_CONTINUE)?.let(::normalizeAppPath),
            )

            else -> return null
        }
    }
}
