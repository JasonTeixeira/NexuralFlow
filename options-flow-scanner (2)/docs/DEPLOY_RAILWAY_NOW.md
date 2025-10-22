# 🚀 DEPLOY TO RAILWAY NOW - STEP BY STEP

**Date**: October 22, 2025  
**Goal**: Deploy WebSocket server to Railway to test real Polygon data  
**Vercel**: Stay on localhost (keep building features)

---

## ✅ WHY THIS APPROACH IS SMART

- ✅ Test infrastructure with REAL Polygon data
- ✅ Verify performance in production
- ✅ Find bugs early
- ✅ Keep building UI locally (no pressure)
- ✅ No users yet (just you testing)

---

## 📋 PREREQUISITES

**You Need:**
1. ✅ Railway account (free tier works)
2. ✅ Polygon API key (from polygon.io)
3. ✅ GitHub account
4. ✅ Git installed locally

**That's it!**

---

## 🚀 DEPLOYMENT STEPS (15 MINUTES)

### **STEP 1: Push Code to GitHub** (3 min)

```bash
# Navigate to project
cd /Users/Sage/Downloads/options-flow-scanner\ \(2\)

# Check status
git status

# Add WebSocket server files
git add websocket-server/
git add docs/
git add lib/services/railway-websocket-client.ts
git add lib/services/redis-cache-service.ts

# Commit
git commit -m "Add production WebSocket server with Polygon integration"

# Push to GitHub
git push origin main
```

**✅ Checkpoint**: Code is on GitHub

---

### **STEP 2: Create Railway Project** (3 min)

1. Go to https://railway.app
2. Click **"Login"** (use GitHub)
3. Click **"New Project"**
4. Select **"Deploy from GitHub repo"**
5. Choose your repository: `options-flow-scanner (2)` or whatever it's called
6. Railway will scan and detect the Dockerfile ✅

**✅ Checkpoint**: Railway project created

---

### **STEP 3: Configure Build Settings** (2 min)

Railway should auto-detect, but verify:

1. In Railway project, click **"Settings"**
2. Under **"Build"**, verify:
   - **Root Directory**: `websocket-server`
   - **Builder**: Dockerfile
   - **Dockerfile Path**: `Dockerfile`
3. Click **"Save"**

**✅ Checkpoint**: Build configured

---

### **STEP 4: Add Environment Variables** (3 min)

1. In Railway project, click **"Variables"**
2. Click **"+ New Variable"**
3. Add these variables:

```
PORT=8080
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3028
POLYGON_API_KEY=your_polygon_api_key_here
```

**How to get Polygon API key:**
- Go to https://polygon.io/dashboard/api-keys
- Copy your key
- Paste above

4. Click **"Add"** for each

**✅ Checkpoint**: Environment variables set

---

### **STEP 5: Add Redis (Optional but Recommended)** (2 min)

1. In Railway project, click **"New"**
2. Select **"Database"**
3. Choose **"Redis"**
4. Railway auto-creates and connects it
5. Environment variables `REDIS_URL` and `REDIS_PASSWORD` auto-set ✅

**✅ Checkpoint**: Redis added

---

### **STEP 6: Deploy** (2 min)

1. Railway automatically starts deploying
2. Watch the logs in real-time
3. Wait for: **"✅ Deployment successful"**

**You should see in logs:**
```
🚀 WebSocket server starting on port 8080
✅ Redis connected successfully
🔌 Connecting to Polygon WebSocket
✅ Connected to Polygon WebSocket
🔐 Authenticating with Polygon...
✅ Polygon authentication successful
✅ Polygon WebSocket integration enabled
📥 Subscribing to Polygon: [SPY QQQ DIA AAPL TSLA NVDA]
✅ Server ready to accept connections
```

**✅ Checkpoint**: Server deployed!

---

### **STEP 7: Get Your WebSocket URL** (1 min)

1. In Railway project, click **"Settings"**
2. Scroll to **"Domains"**
3. Click **"Generate Domain"**
4. You'll get: `your-app.up.railway.app`

**Your WebSocket URL is:**
```
wss://your-app.up.railway.app/ws
```

**Copy this!** You'll need it.

**✅ Checkpoint**: URL obtained

---

## 🧪 TEST YOUR DEPLOYMENT (5 MIN)

### **Test 1: Health Check**

```bash
curl https://your-app.up.railway.app/health
```

**Expected:**
```json
{
  "status": "healthy",
  "clients": 0,
  "uptime": 123.45,
  "redis": true,
  "timestamp": 1698345678
}
```

✅ **If you see this, server is running!**

---

### **Test 2: WebSocket Connection**

```bash
# Install wscat (one time)
npm install -g wscat

# Connect to your server
wscat -c wss://your-app.up.railway.app/ws
```

**You should see:** `Connected`

**Now subscribe to AAPL:**
```json
{"type":"subscribe","channel":"trades","symbols":["AAPL"]}
```

**You should see:**
```json
{"type":"subscribed","channel":"trades","symbols":["AAPL"],"timestamp":1698345678123}
```

**Then real-time trades will stream in!** 🎉
```json
{
  "type":"market-data",
  "channel":"trades",
  "data":{
    "ev":"T",
    "sym":"AAPL",
    "p":150.25,
    "s":100,
    "x":4,
    "t":1698345678000
  },
  "timestamp":1698345678123,
  "symbols":["AAPL"]
}
```

**✅ If you see trades streaming, IT'S WORKING!** 🚀

---

### **Test 3: Connect from Your Local Next.js App**

1. **Update `.env.local` on your machine:**

```bash
# Add this line
NEXT_PUBLIC_RAILWAY_WS_URL=wss://your-app.up.railway.app/ws
```

2. **Restart your Next.js dev server:**

```bash
npm run dev
```

3. **Your local app now connects to Railway!**

The WebSocket client will automatically use your Railway URL.

**✅ Checkpoint**: Local app connected to Railway server

---

## 📊 MONITORING

### **View Logs**

In Railway dashboard:
- Click **"Deployments"**
- Click **"View Logs"**
- Watch real-time logs

**Or use CLI:**
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Link project
railway link

# View logs
railway logs
```

---

### **Check Stats**

```bash
curl https://your-app.up.railway.app/stats
```

**Shows:**
- Connected clients
- Active channels
- Uptime
- Redis status

---

## 💰 COST

**Free Tier:**
- $5/month credit
- Good for testing
- 500 hours/month

**Hobby Plan:**
- $5/month base
- $0.000231/GB-hour for usage
- **Estimated: $10-15/month for your server**

**With Redis:**
- +$5-10/month

**Total: $15-25/month** ✅

---

## ⚠️ IMPORTANT NOTES

### **Keep Vercel on Localhost**

**DON'T deploy to Vercel yet** because:
- ❌ No user authentication
- ❌ No database persistence
- ❌ No payment system
- ❌ UI not polished

**Deploy Vercel when:**
- ✅ Supabase integrated
- ✅ Auth system working
- ✅ Payments configured
- ✅ UI polished
- ✅ Ready for users

### **Your Workflow Now**

```bash
# 1. Code features locally (Vercel on localhost:3000)
npm run dev

# 2. Your local app connects to Railway WebSocket
#    (using NEXT_PUBLIC_RAILWAY_WS_URL)

# 3. Test with REAL Polygon data

# 4. When happy, update WebSocket server:
cd websocket-server
# Make changes
git add .
git commit -m "Update WebSocket server"
git push origin main
# Railway auto-deploys in 2 min

# 5. Keep coding locally on Vercel
```

---

## 🐛 TROUBLESHOOTING

### **Issue: Deployment Failed**

**Solution:**
1. Check Railway logs for errors
2. Verify Dockerfile is in `websocket-server/`
3. Verify Root Directory is set to `websocket-server`

### **Issue: Can't Connect to WebSocket**

**Solution:**
1. Check ALLOWED_ORIGINS includes `http://localhost:3000`
2. Verify WebSocket URL uses `wss://` not `ws://`
3. Check Railway logs for connection errors

### **Issue: Polygon Authentication Failed**

**Solution:**
1. Verify POLYGON_API_KEY is correct
2. Check you have an active Polygon subscription
3. Test API key at polygon.io dashboard

### **Issue: High Costs**

**Solution:**
1. Railway free tier gives $5/month credit
2. Monitor usage in Railway dashboard
3. Hobby plan is ~$15-25/month total
4. You can pause/stop deployment anytime

---

## ✅ SUCCESS CHECKLIST

After deployment, verify:

- [ ] Railway project created
- [ ] Code pushed to GitHub
- [ ] Environment variables set
- [ ] Redis added (optional)
- [ ] Deployment successful
- [ ] Health check returns 200
- [ ] WebSocket connection works
- [ ] Can subscribe to symbols
- [ ] Real-time trades streaming
- [ ] Local Next.js app connects
- [ ] Logs show no errors

**When all checked, you're LIVE!** 🎉

---

## 🎯 WHAT'S NEXT

### **After Railway is Deployed**

1. **Test with real data** (ongoing)
2. **Monitor performance** (Railway dashboard)
3. **Build Supabase integration** (next)
4. **Add authentication** (after Supabase)
5. **Deploy Vercel** (when ready for users)

### **Timeline to Full Production**

- ✅ **Today**: Railway deployed
- Week 1: Supabase + Auth
- Week 2: Data persistence + Stripe
- Week 3: Polish UI + Testing
- Week 4: Deploy Vercel + Go live

---

## 🚀 READY TO DEPLOY?

**Just follow the 7 steps above!**

**Questions?**
- Railway docs: https://docs.railway.app
- Polygon docs: https://polygon.io/docs
- wscat: `npm install -g wscat`

**This deployment won't affect your local development.**  
**You can keep coding while Railway handles the infrastructure.** ✅

**LET'S DO THIS!** 🔥
