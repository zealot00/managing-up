import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const publicPaths = ["/", "/login", "/api/v1/auth/login", "/api/v1/auth/logout"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Allow public paths
  if (publicPaths.some((path) => pathname === path || pathname.startsWith("/api/"))) {
    return NextResponse.next();
  }
  
  // Check for auth token cookie
  const authToken = request.cookies.get("auth_token");
  if (!authToken) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|logo.svg).*)"],
};
