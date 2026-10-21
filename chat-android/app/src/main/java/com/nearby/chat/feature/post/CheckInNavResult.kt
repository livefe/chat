package com.nearby.chat.feature.post

data class CheckInNavResult(
    val payloadJson: String,
    val kind: String,
) {
    companion object {
        const val KIND_SUCCESS = "success"
        const val KIND_ALREADY = "already"
        const val KIND_FAIL = "fail"
    }
}
