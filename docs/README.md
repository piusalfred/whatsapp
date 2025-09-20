# Get Started Guide

## Initial Steps

1. **Review WhatsApp Cloud API Documentation**  
   Read the official [WhatsApp Cloud API Get Started Guide](https://developers.facebook.com/docs/whatsapp/cloud-api/get-started) to understand the requirements and messaging rules.

2. **Set Up a Facebook Developer App**
    - Go to the [Facebook Developer Console](https://developers.facebook.com/apps).
    - Select an existing app, or create a new one.
    - Under your app, navigate to the **WhatsApp > API Setup** section.

3. **Obtain Required Credentials**  
   From the app dashboard, collect the following:
    - `access_token`
    - `phone_number_id`
    - `business_account_id`  
      Ensure that you've authorized one or more phone numbers for message sending.

4. **Build Example Binaries**  
   Run the following from the project root to compile example binaries into `examples/bin`:

   ```bash
   task build-examples
   ```

5. **Configure Environment Variables**  
   Copy the provided environment template:

   ```bash
   cp examples/api.env examples/bin/api.env
   ```
   Then edit examples/bin/api.env and replace the placeholders with your actual credentials:

    ```dotenv
    WHATSAPP_CLOUD_API_BASE_URL=https://graph.facebook.com
    WHATSAPP_CLOUD_API_API_VERSION=v20.0
    WHATSAPP_CLOUD_API_ACCESS_TOKEN=your_access_token_here
    WHATSAPP_CLOUD_API_PHONE_NUMBER_ID=your_phone_number_id_here
    WHATSAPP_CLOUD_API_BUSINESS_ACCOUNT_ID=your_business_account_id_here
    WHATSAPP_CLOUD_API_TEST_RECIPIENT=+1234567890
    ```
6. **Run the Examples**  
   Navigate to the examples directory:

   ```bash
   cd examples/bin
   ```
   > Before running any examples, ensure that you have sent the "Hello World" template message to your test recipient and received a reply within the 24-hour messaging window.