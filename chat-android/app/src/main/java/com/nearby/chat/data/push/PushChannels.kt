package com.nearby.chat.data.push

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import com.nearby.chat.R

object PushChannels {
    const val PRIVATE_CHAT = "private_chat"

    fun ensure(context: Context) {
        val mgr = context.getSystemService(NotificationManager::class.java) ?: return
        if (mgr.getNotificationChannel(PRIVATE_CHAT) != null) return
        mgr.createNotificationChannel(
            NotificationChannel(
                PRIVATE_CHAT,
                context.getString(R.string.push_channel_private_name),
                NotificationManager.IMPORTANCE_HIGH,
            ).apply {
                description = context.getString(R.string.push_channel_private_description)
            },
        )
    }
}
