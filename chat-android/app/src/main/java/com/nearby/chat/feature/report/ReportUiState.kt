package com.nearby.chat.feature.report

data class ReportUiState(
    val reason: Int = ReportReason.SPAM.value,
    val content: String = "",
    val images: List<ReportImageItem> = emptyList(),
    val loading: Boolean = false,
) {
    val hasUploadingImages: Boolean get() = images.any { it.uploading }
}
