import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
 
// This function can be marked `async` if using `await` inside
export function middleware(request: NextRequest) {
  console.log('Middleware executed', request.nextUrl.pathname)
  if (protectedRoutes.includes(request.nextUrl.pathname) ) {
    console.log('Protected route detected')
    const token = request.cookies.get('access_token')
    if (!token) {
      return NextResponse.redirect(new URL('/login', request.url))
    }
    console.log('Token found')
    return NextResponse.next()
  }
  console.log('Unprotected route detected')
  return NextResponse.next()
}

export const protectedRoutes = ['/dashboard']
export const config = {
  matcher: ['/dashboard/:path*'],
}