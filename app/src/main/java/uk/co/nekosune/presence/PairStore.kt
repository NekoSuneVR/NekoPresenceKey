package uk.co.nekosune.presence

import android.content.Context

object PairStore {
    private const val PREFS = "nekopresence_pair"

    fun save(context: Context, host: String, pairId: String, computerKey: String) {
        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE).edit()
            .putString("host", host)
            .putString("pair_id", pairId)
            .putString("computer_key", computerKey)
            .apply()
    }

    fun host(context: Context): String? = prefs(context).getString("host", null)
    fun pairId(context: Context): String? = prefs(context).getString("pair_id", null)
    fun computerKey(context: Context): String? = prefs(context).getString("computer_key", null)
    fun paired(context: Context): Boolean = host(context) != null && pairId(context) != null && computerKey(context) != null

    private fun prefs(context: Context) = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
}
