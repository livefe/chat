package com.nearby.chat.feature.feedback

data class FeedbackUiState(
    val type: Int = FeedbackType.SUGGESTION.value,
    val content: String = "",
    val contact: String = "",
    val images: List<FeedbackImageItem> = emptyList(),
    val loading: Boolean = false,
) {
    val hasUploadingImages: Boolean get() = images.any { it.uploading }
}
