// ===================== SUPABASE =====================
const SUPABASE_URL = "https://xkzadsnvjdrspymgcoaq.supabase.co";
const SUPABASE_KEY = "sb_publishable_YziaIyx4gzbIUUm3lSbIXw_eiltdzur";

// async function loadEnv() {
//   try {
//     const res = await fetch(".env");
//     if (!res.ok) throw new Error("Failed to load .env");
//     const raw = await res.text();
//     raw.split(/\r?\n/).forEach((line) => {
//       const trimmed = line.trim();
//       if (!trimmed || trimmed.startsWith("#")) return;
//       const [key, ...rest] = trimmed.split("=");
//       const value = rest.join("=").trim();
//       if (key === "SUPABASE_URL") SUPABASE_URL = value;
//       if (key === "SUPABASE_KEY") SUPABASE_KEY = value;
//     });
//   } catch (e) {
//     console.error("Env load error:", e);
//     throw e;
//   }
// }

    async function sb(
        table,
        method = "GET",
        body = null,
        match = null,
        select = null,
    ) {
        let url = `${SUPABASE_URL}/rest/v1/${table}`;
        const params = [];
        if (select) params.push(`select=${select}`);
        if (match)
            Object.entries(match).forEach(([k, v]) => params.push(`${k}=eq.${v}`));
        if (params.length) url += "?" + params.join("&");

        const headers = {
            apikey: SUPABASE_KEY,
            Authorization: `Bearer ${sbToken || SUPABASE_KEY}`,
            "Content-Type": "application/json",
            Prefer:
                method === "POST" ? "return=minimal" : "return=representation",
        };
        if (method === "GET") headers["Accept"] = "application/json";

        const res = await fetch(url, {
            method,
            headers,
            body: body ? JSON.stringify(body) : null,
        });
        if (!res.ok) {
            const err = await res.text();
            throw new Error(err);
        }
        const text = await res.text();
        return text ? JSON.parse(text) : [];
    }

    async function sbAuth(email, password) {
        const res = await fetch(`${SUPABASE_URL}/auth/v1/token?grant_type=password`, {
            method: "POST",
            headers: {
                apikey: SUPABASE_KEY,
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ email, password }),
        });
        const data = await res.json();
        if (!res.ok)
            throw new Error(data.error_description || data.msg || "Login failed");
        return data.access_token;
    }

    async function sbLogout() {
        if (!sbToken) return;
        await fetch(`${SUPABASE_URL}/auth/v1/logout`, {
            method: "POST",
            headers: {
                apikey: SUPABASE_KEY,
                Authorization: `Bearer ${sbToken}`,
            },
        });
        sbToken = null;
    }

    // ===================== PAGE VIEW TRACKING =====================
    async function trackVisit() {
        // Deduplicate: only track once per session
        if (sessionStorage.getItem("pv_tracked")) return;
        try {
            await fetch(`${SUPABASE_URL}/rest/v1/page_views`, {
                method: "POST",
                headers: {
                    apikey: SUPABASE_KEY,
                    Authorization: `Bearer ${SUPABASE_KEY}`,
                    "Content-Type": "application/json",
                    Prefer: "return=minimal",
                },
                body: JSON.stringify({ page: "home" }),
            });
            sessionStorage.setItem("pv_tracked", "1");
        } catch (e) {
            /* fail silently — never break the site for visitors */
        }
    }