import { defineLoader } from 'vitepress'
import { createSign } from 'node:crypto'
import { readFileSync, existsSync } from 'node:fs'

export interface Review {
  title: string
  body: string
  author: string
}

declare const data: Review[]
export { data }

const KEY_ID = 'T43ZQ8KTCH'
const ISSUER_ID = '1c165417-1b30-448f-bfab-acfd9e96747a'
const KEY_FILE = '/home/ondrej/projects/diction/.secrets/AuthKey_T43ZQ8KTCH.p8'
const APP_ID = '6759807364'
const ENGLISH_TERRITORIES = new Set(['USA', 'GBR', 'CAN', 'AUS', 'NZL', 'IRL', 'LTU'])

function makeJWT(): string {
  const privateKey = readFileSync(KEY_FILE, 'utf-8')
  const header = Buffer.from(JSON.stringify({ alg: 'ES256', kid: KEY_ID, typ: 'JWT' })).toString('base64url')
  const now = Math.floor(Date.now() / 1000)
  const payload = Buffer.from(JSON.stringify({
    iss: ISSUER_ID, iat: now, exp: now + 1200, aud: 'appstoreconnect-v1'
  })).toString('base64url')
  const unsigned = `${header}.${payload}`
  const sign = createSign('SHA256')
  sign.update(unsigned)
  const sig = sign.sign({ key: privateKey, dsaEncoding: 'ieee-p1363' }, 'base64url')
  return `${unsigned}.${sig}`
}

export default defineLoader({
  async load(): Promise<Review[]> {
    try {
      if (!existsSync(KEY_FILE)) {
        console.warn('[reviews] ASC key not found, skipping')
        return []
      }
      const token = makeJWT()
      const url = new URL(`https://api.appstoreconnect.apple.com/v1/apps/${APP_ID}/customerReviews`)
      url.searchParams.set('limit', '50')
      url.searchParams.set('sort', '-createdDate')
      const res = await fetch(url.toString(), { headers: { Authorization: `Bearer ${token}` } })
      const json = await res.json() as any
      return (json.data ?? [])
        .filter((r: any) => r.attributes.rating === 5 && ENGLISH_TERRITORIES.has(r.attributes.territory))
        .map((r: any) => ({
          title: r.attributes.title,
          body: r.attributes.body,
          author: r.attributes.reviewerNickname,
        }))
    } catch (e) {
      console.warn('[reviews] Failed:', e)
      return []
    }
  },
})
