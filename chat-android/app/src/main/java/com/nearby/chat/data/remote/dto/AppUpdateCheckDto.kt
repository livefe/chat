package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class AppUpdateCheckDto(
    val latestVersion: String = "",
    val minSupportedVersion: String = "",
    val downloadUrl: String = "",
    val fileSize: Long = 0,
    val hasUpdate: Boolean = false,
    val forceUpdate: Boolean = false,
    val title: String = "",
    val message: String = "",
)
