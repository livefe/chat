import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.hilt)
    alias(libs.plugins.ksp)
}

val localProps =
    Properties().apply {
        val f = rootProject.file("local.properties")
        if (f.exists()) {
            f.inputStream().use { load(it) }
        }
    }

fun localProp(
    key: String,
    default: String = "",
): String = (localProps.getProperty(key) ?: default).replace("\\", "\\\\").replace("\"", "\\\"")

fun shareLinkHost(): String {
    val raw = localProp("H5_BASE_URL").trim()
    if (raw.isBlank()) return ""
    val withoutScheme = raw.substringAfter("://", raw)
    return withoutScheme.substringBefore('/').substringBefore(':')
}

fun shareLinkScheme(): String {
    val raw = localProp("H5_BASE_URL").trim()
    if (raw.isBlank()) return "https"
    return raw.substringBefore("://", missingDelimiterValue = "https").ifBlank { "https" }
}

fun isLanShareHost(host: String): Boolean {
    if (host.isBlank() || host.equals("localhost", ignoreCase = true)) return true
    return host.matches(Regex("""^\d{1,3}(\.\d{1,3}){3}$"""))
}

fun appLinkManifestHost(): String {
    val host = shareLinkHost()

    if (isLanShareHost(host)) return "app-links.disabled.local"
    return host.ifBlank { "app-links.disabled.local" }
}

android {
    namespace = "com.nearby.chat"
    compileSdk =
        libs.versions.compileSdk
            .get()
            .toInt()

    defaultConfig {
        applicationId = "com.nearby.chat"
        minSdk =
            libs.versions.minSdk
                .get()
                .toInt()
        targetSdk =
            libs.versions.targetSdk
                .get()
                .toInt()
        versionCode = 1
        versionName = "1.0.0"

        buildConfigField("String", "H5_BASE_URL", "\"${localProp("H5_BASE_URL")}\"")
        buildConfigField("String", "SHARE_LINK_HOST", "\"${shareLinkHost()}\"")
        buildConfigField("String", "SHARE_LINK_SCHEME", "\"${shareLinkScheme()}\"")
        buildConfigField("String", "TIANDITU_TK", "\"${localProp("TIANDITU_TK")}\"")
        buildConfigField("String", "XIAOMI_APP_ID", "\"${localProp("XIAOMI_APP_ID")}\"")
        buildConfigField("String", "XIAOMI_APP_KEY", "\"${localProp("XIAOMI_APP_KEY")}\"")

        manifestPlaceholders["shareLinkHost"] = appLinkManifestHost()
        manifestPlaceholders["shareLinkScheme"] = shareLinkScheme()

        vectorDrawables.useSupportLibrary = true
    }

    buildTypes {
        debug {
            isDebuggable = true

            val debugApiDefault = "http://10.0.2.2/"
            buildConfigField("String", "API_BASE_URL", "\"${localProp("API_BASE_URL", debugApiDefault)}\"")
        }
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            buildConfigField("String", "API_BASE_URL", "\"${localProp("API_BASE_URL", "")}\"")
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro",
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }
}

dependencies {
    val mipushAars =
        fileTree("libs") {
            include("*.aar")
        }
    if (!mipushAars.files.isEmpty()) {
        implementation(mipushAars)
    } else {
        implementation(project(":mipush-stub"))
    }

    val composeBom = platform(libs.androidx.compose.bom)
    implementation(composeBom)

    implementation(libs.androidx.core.ktx)
    implementation(libs.zxing.core)
    implementation(libs.zxing.android.embedded)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)

    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material.icons.extended)
    debugImplementation(libs.androidx.compose.ui.tooling)

    implementation(libs.androidx.navigation.compose)
    implementation(libs.androidx.hilt.navigation.compose)

    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)

    implementation(libs.retrofit)
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging.interceptor)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.retrofit.kotlinx.serialization.converter)

    implementation(libs.androidx.datastore.preferences)
    implementation(libs.androidx.security.crypto)

    implementation(libs.coil.compose)
    implementation(libs.coil.video)

    implementation(libs.accompanist.permissions)

    implementation(libs.androidx.browser)

    implementation(libs.play.services.location)
    implementation(libs.kotlinx.coroutines.play.services)

    implementation(libs.osmdroid.android)
}
