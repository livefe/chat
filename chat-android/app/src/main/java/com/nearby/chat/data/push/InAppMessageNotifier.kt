package com.nearby.chat.data.push

import android.app.PendingIntent
import android.content.Context
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.nearby.chat.R
import com.nearby.chat.core.lifecycle.AppVisibility
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class InAppMessageNotifier @Inject constructor(
    @ApplicationContext private val context: Context,
    private val appVisibility: AppVisibility,
) {
    fun notifyPrivate(peerId: Long, title: String, body: String) {
        if (peerId <= 0L) return

        if (appVisibility.isForeground.value) return

        PushChannels.ensure(context)
        val intent = PushIntents.openPrivateChat(context, peerId) ?: return
        val pending = PendingIntent.getActivity(
            context,
            peerId.toInt(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        val notification = NotificationCompat.Builder(context, PushChannels.PRIVATE_CHAT)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(title.ifBlank { context.getString(R.string.notification_new_message_title) })
            .setContentText(body.ifBlank { context.getString(R.string.notification_new_message_body) })
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setAutoCancel(true)
            .setContentIntent(pending)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setCategory(NotificationCompat.CATEGORY_MESSAGE)
            .build()
        runCatching {
            NotificationManagerCompat.from(context).notify(NOTIF_BASE + peerId.toInt(), notification)
        }
    }

    fun cancelPrivate(peerId: Long) {
        if (peerId <= 0L) return
        NotificationManagerCompat.from(context).cancel(NOTIF_BASE + peerId.toInt())
    }

    companion object {
        private const val NOTIF_BASE = 41_000
    }
}
