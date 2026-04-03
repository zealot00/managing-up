import {NextRequest, NextResponse} from 'next/server';
import {match as matchLocale} from '@formatjs/intl-localematcher';
import Negotiator from 'negotiator';

const locales = ['en', 'zh'];
const defaultLocale = 'en';

const publicPaths = ['/', '/login', '/api/v1/auth/login', '/api/v1/auth/logout'];

function getLocale(request: NextRequest): string {
  const cookieLocale = request.cookies.get('NEXT_LOCALE')?.value;
  if (cookieLocale && locales.includes(cookieLocale)) {
    return cookieLocale;
  }

  const negotiatorHeaders: Record<string, string> = {};
  request.headers.forEach((value, key) => {
    negotiatorHeaders[key] = value;
  });

  const languages = new Negotiator({headers: negotiatorHeaders}).languages();
  return matchLocale(languages, locales, defaultLocale);
}

export function middleware(request: NextRequest) {
  const {pathname} = request.nextUrl;

  // Skip API routes, static files, and _next
  if (
    pathname.startsWith('/api/') ||
    pathname.startsWith('/_next/') ||
    pathname.startsWith('/favicon.ico') ||
    pathname.startsWith('/logo.svg')
  ) {
    return NextResponse.next();
  }

  // Auth check
  if (!publicPaths.some((path) => pathname === path || pathname.startsWith('/api/'))) {
    const authToken = request.cookies.get('auth_token');
    if (!authToken) {
      return NextResponse.redirect(new URL('/login', request.url));
    }
  }

  // Set locale cookie for client-side access
  const locale = getLocale(request);
  const response = NextResponse.next();
  response.cookies.set('NEXT_LOCALE', locale, {
    path: '/',
    maxAge: 365 * 24 * 60 * 60,
    sameSite: 'lax',
  });

  return response;
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico|logo.svg).*)'],
};
