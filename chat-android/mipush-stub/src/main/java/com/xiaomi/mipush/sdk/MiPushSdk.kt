package com.xiaomi.mipush.sdk

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

object MiPushClient {
    const val COMMAND_REGISTER = "register"

    @JvmStatic
    fun setRegion(region: Region) {
    }

    @JvmStatic
    fun registerPush(context: Context, appId: String, appKey: String) {
    }

    @JvmStatic
    fun unregisterPush(context: Context) {
    }

    @JvmStatic
    fun getRegId(context: Context): String? = null
}

enum class Region {
    China,
    Global,
    Europe,
    Russia,
    India,
}

abstract class PushMessageReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {}

    open fun onReceivePassThroughMessage(context: Context, message: MiPushMessage) {}
    open fun onNotificationMessageClicked(context: Context, message: MiPushMessage) {}
    open fun onNotificationMessageArrived(context: Context, message: MiPushMessage) {}
    open fun onCommandResult(context: Context, message: MiPushCommandMessage) {}
    open fun onReceiveRegisterResult(context: Context, message: MiPushCommandMessage) {}
    open fun onRequirePermissions(context: Context, permissions: Array<String>) {}
}

class MiPushMessage {
    private var content: String? = null
    private var title: String? = null
    private var description: String? = null
    private var extra: MutableMap<String, String> = linkedMapOf()
    private var topic: String? = null
    private var alias: String? = null
    private var userAccount: String? = null

    fun getContent(): String? = content
    fun getTitle(): String? = title
    fun getDescription(): String? = description
    fun getExtra(): MutableMap<String, String> = extra
    fun getTopic(): String? = topic
    fun getAlias(): String? = alias
    fun getUserAccount(): String? = userAccount

    fun setContent(v: String?) {
        content = v
    }

    fun setExtra(v: MutableMap<String, String>) {
        extra = v
    }
}

class MiPushCommandMessage {
    private var command: String? = null
    private var resultCode: Long = -1
    private var reason: String? = null
    private var commandArguments: List<String>? = null

    fun getCommand(): String? = command
    fun getResultCode(): Long = resultCode
    fun getReason(): String? = reason
    fun getCommandArguments(): List<String>? = commandArguments

    fun setCommand(v: String?) {
        command = v
    }

    fun setResultCode(v: Long) {
        resultCode = v
    }

    fun setCommandArguments(v: List<String>?) {
        commandArguments = v
    }
}

object ErrorCode {
    const val SUCCESS: Long = 0
}
