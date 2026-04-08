import {ReactNode} from 'react';
import {NextIntlClientProvider} from 'next-intl';
import {getMessages} from 'next-intl/server';
import {ToastProvider} from '../components/ToastProvider';
import {QueryProvider} from './components/providers/QueryProvider';

export default async function Providers({children}: {children: ReactNode}) {
  const messages = await getMessages();

  return (
    <NextIntlClientProvider messages={messages}>
      <QueryProvider>
        <ToastProvider>
          {children}
        </ToastProvider>
      </QueryProvider>
    </NextIntlClientProvider>
  );
}