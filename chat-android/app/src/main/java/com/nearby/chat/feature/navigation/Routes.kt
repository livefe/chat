package com.nearby.chat.feature.navigation

import android.net.Uri

object Routes {
    const val LOGIN = "login"
    const val H5_WEB = "h5Web?path={path}&title={title}"
    const val MAIN = "main"
    const val PRIVATE_CHAT = "privateChat/{peerId}?name={name}"
    const val SUPPORT_CHAT = "supportChat"
    const val USER_PROFILE = "userProfile/{userId}"
    const val ROOM = "room/{roomId}"
    const val NEARBY_ROOMS = "nearbyRooms"
    const val POST_DETAIL = "postDetail/{postId}"
    const val EDIT_PROFILE = "editProfile"
    const val EDIT_NICKNAME = "editNickname"
    const val EDIT_MOBILE = "editMobile"
    const val FEEDBACK = "feedback"
    const val FEEDBACK_RECORDS = "feedbackRecords"
    const val FEEDBACK_DETAIL = "feedbackDetail/{id}"
    const val SETTINGS = "settings"
    const val BLOCK_LIST = "blockList"
    const val ACCOUNT_DEACTIVATE = "accountDeactivate"
    const val MY_POSTS = "myPosts/{mode}"
    const val POST_SIGNUPS = "postSignups/{postId}"
    const val POST_CHECK_IN_CODE = "postCheckInCode/{postId}"
    const val CHECK_IN_SCAN = "checkInScan"
    const val CHECK_IN_RESULT = "checkInResult?payload={payload}"
    const val REPORT = "report?targetType={targetType}&targetId={targetId}&title={title}&avatar={avatar}"
    const val ROOM_CREATE = "roomCreate"
    const val ROOM_APPLICATIONS = "roomApplications"
    const val POST_CREATE = "postCreate?roomId={roomId}&postId={postId}"
    const val LOCATION_PICK = "locationPick?for={for}"
    const val FORWARD = "forward"

    const val MY_POSTS_PUBLISHED = "published"
    const val MY_POSTS_JOINED = "joined"
    const val PICK_FOR_SHARE = "share"

    const val H5_AGREEMENT_USER = "agreement/user"
    const val H5_AGREEMENT_PRIVACY = "agreement/privacy"

    fun privateChat(peerId: Long, name: String = "") = "privateChat/$peerId?name=${Uri.encode(name)}"

    fun userProfile(userId: Long) = "userProfile/$userId"
    fun room(roomId: Long) = "room/$roomId"
    fun postDetail(postId: Long) = "postDetail/$postId"
    fun postCreate(roomId: Long = 0, postId: Long = 0) = "postCreate?roomId=$roomId&postId=$postId"
    fun postSignups(postId: Long) = "postSignups/$postId"
    fun postCheckInCode(postId: Long) = "postCheckInCode/$postId"
    fun checkInScan() = "checkInScan"
    fun checkInResult(payload: String) = "checkInResult?payload=${Uri.encode(payload)}"
    fun report(targetType: Int, targetId: Long, title: String = "", avatar: String = "") = "report?targetType=$targetType&targetId=$targetId" +
        "&title=${Uri.encode(title)}" +
        "&avatar=${Uri.encode(avatar)}"
    fun myPosts(mode: String) = "myPosts/$mode"
    fun feedbackDetail(id: Long) = "feedbackDetail/$id"

    fun h5Web(path: String, title: String = "") = "h5Web?path=${Uri.encode(path)}&title=${Uri.encode(title)}"
    fun locationPick(purpose: String = "") = "locationPick?for=${Uri.encode(purpose)}"
}
