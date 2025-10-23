# Railway REST API Deployment

## Changes Made

1. **Added REST API handlers** (`api-handlers.go`)
   - Portfolio summary endpoint
   - Watchlist endpoint
   - Market pulse endpoint
   - Portfolio snapshot endpoint
   - Today's opportunities endpoint

2. **Registered routes in main.go**
   - `/api/portfolio/summary`
   - `/api/watchlist`
   - `/api/market/pulse`
   - `/api/portfolio/snapshot`
   - `/api/opportunities/today`

3. **Features**
   - Redis caching (60-second TTL)
   - CORS enabled
   - Fast response times
   - Mock data (replace with real Polygon calls in production)

## Deploy to Railway

```bash
# From project root
cd websocket-server

# Add files
git add api-handlers.go main.go DEPLOY.md

# Commit
git commit -m "Add REST API endpoints for dashboard"

# Push (Railway auto-deploys)
git push origin main
```

## Test Endpoints

Once deployed, test with:

```bash
# Replace with your Railway URL
RAILWAY_URL="https://nexural-flow-production.up.railway.app"

# Test portfolio summary
curl $RAILWAY_URL/api/portfolio/summary

# Test watchlist
curl $RAILWAY_URL/api/watchlist

# Test market pulse
curl $RAILWAY_URL/api/market/pulse

# Test health
curl $RAILWAY_URL/health
```

## Frontend Integration

Update `.env.local` in Next.js project:

```bash
# Add this line
NEXT_PUBLIC_API_BASE_URL=https://nexural-flow-production.up.railway.app
```

Then restart Next.js:
```bash
npm run dev
```

The dashboard will now fetch from Railway instead of local APIs!
