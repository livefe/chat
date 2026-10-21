package com.nearby.chat.feature.post

import android.net.Uri
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import com.nearby.chat.core.model.CheckInResultPayload
import com.nearby.chat.core.post.decodeCheckInResultPayload
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class PostCheckInResultViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    val payload: CheckInResultPayload? = decodeCheckInResultPayload(
        Uri.decode(savedStateHandle.get<String>(ARG_PAYLOAD).orEmpty()),
    )
    val postId: Long = payload?.postId ?: 0L

    companion object {
        const val ARG_PAYLOAD = "payload"
    }
}
