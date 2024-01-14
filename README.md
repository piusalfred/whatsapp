# whatsapp


## set up
- Have golang installed
- Register as a Meta Developer here https://developers.facebook.com/docs/development/register 
- Create an application here https://developers.facebook.com/docs/development/create-an-app and configure
it to enable access to WhatsApp Business Cloud API and Webhooks.

- You can manage your apps here https://developers.facebook.com/apps/

- From Whatsapp Developer Dashboard you can try and send a test message to your phone number.
to be sure that everything is working fine before you start using this api. Also you need to
reply to that message to be able to send other messages.

- Go to [examples/base](examples/base) then create `.env` file that looks like [examples/base/.envrc](examples/base/.envrc)
 and add your credentials there.

- Run `make run` and wait to receive a message on your phone. Make sure you have sent the template message
first from the Whatsapp Developer Dashboard.
