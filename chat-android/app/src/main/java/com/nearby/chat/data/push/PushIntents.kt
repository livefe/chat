package com.nearby.chat.data.push

import android.content.Context
import android.content.Intent
import com.nearby.chat.core.share.ShareLinks

object PushIntents {
    fun openPrivateChat(context: Context, peerId: Long): Intent? = context.packageManager.getLaunchIntentForPackage(context.packageName)?.apply {
        flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        if (peerId > 0) putExtra(ShareLinks.EXTRA_PEER_ID, peerId)
    }
}
