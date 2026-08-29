package uk.co.nekosune.presence

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import java.net.InetAddress

object NetworkGuard {
    fun isTrustedTransport(context: Context): Boolean {
        val cm = context.getSystemService(ConnectivityManager::class.java)
        val network = cm.activeNetwork ?: return false
        val caps = cm.getNetworkCapabilities(network) ?: return false
        return caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) &&
            !caps.hasTransport(NetworkCapabilities.TRANSPORT_VPN)
    }

    fun isPrivateHost(host: String): Boolean = try {
        val ip = InetAddress.getByName(host)
        ip.isSiteLocalAddress || ip.isLinkLocalAddress
    } catch (_: Exception) { false }
}
