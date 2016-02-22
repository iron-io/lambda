import java.lang.reflect.Method;
import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;
import java.util.Objects;

import com.amazonaws.services.lambda.runtime.ClientContext;
import com.amazonaws.services.lambda.runtime.CognitoIdentity;
import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.LambdaLogger;

import com.google.gson.Gson;

public class LambdaLauncher {
    public static void main(String[] args) {
        String handler_env = System.getenv("handler");
        String payload = System.getenv("payload");

        if (handler_env == null) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        if (payload == null || payload.equals("")) {
            System.out.println("Payload is empty, please specify payload");
            System.exit(1);
        }

        String[] package_function = handler_env.split("::");
        if (package_function[0] == null || package_function[0].equals("") || package_function[1] == null || package_function[1].equals("")) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        try {
            LambdaLauncher ll = new LambdaLauncher();
            ll.launch_function(package_function, payload);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    public void launch_function(String[] packageFunction, String payload) throws Exception {
        String payloadType;
        boolean processed = false;
        boolean classFound = false;
        System.out.println("PAYLOAD: " + payload);
        if (PayloadHelper.isInteger(payload)) {
            payloadType = "int";
            System.out.println("PAYLOAD IS INTEGER!");
        } else if (PayloadHelper.isJSONValid(payload)) {
            payloadType = "json";
            System.out.println("PAYLOAD IS JSON!");
        } else {
            payloadType = "string";
            System.out.println("PAYLOAD IS String!");
        }

        AWSContext aws_ctx = new AWSContext();

        Class cls = Class.forName(packageFunction[0]); //get class in package
        Object lambdaClass = cls.newInstance();

        Method[] declaredMethods = cls.getDeclaredMethods();
        for (Method method : declaredMethods) {
            if (Objects.equals(method.getName(), packageFunction[1])) {  //if method name in user class == method name in env var
                classFound = true;
                System.out.println("Found package: " + packageFunction[0] + " and method: " + packageFunction[1]);

                Class[] parameterTypes = method.getParameterTypes();

                Method lambdaMethod = cls.getDeclaredMethod(packageFunction[1], parameterTypes);
                Object result = null;

                //TODO: count and check params, first and second maybe string,int,pojo(class), second or third must be context, first and second maybe input/output stream
                for (Class parameterType : parameterTypes) {

                    if (Objects.equals(payloadType, "int") && (Objects.equals(parameterType.toString(), payloadType) || Objects.equals(parameterType.toString(), "Integer"))) {
                        result = lambdaMethod.invoke(lambdaClass, Integer.parseInt(payload), aws_ctx);
                        processed = true;
                    } else if (Objects.equals(payloadType, "string") && (Objects.equals(parameterType.toString(), payloadType) || Objects.equals(parameterType.toString(), "String"))) {
                        result = lambdaMethod.invoke(lambdaClass, payload, aws_ctx);
                        processed = true;
                    } else if (Objects.equals(payloadType, "json")) {
                        if (parameterType.getName().toLowerCase().contains("RequestClass".toLowerCase())) { //TODO
                            result = PayloadHelper.gson.toJson(lambdaMethod.invoke(lambdaClass, PayloadHelper.gson.fromJson(payload, parameterType), aws_ctx));
                            processed = true;
                        }
                    }
                }

                String ll_result = processed ? String.format("Method executed with result: %s", (String) result) :
                        String.format("Handler %s with param of %s type not found", packageFunction[1], payloadType);
                System.out.println(ll_result);
            }
        }
        if (!classFound) System.out.println(String.format("Class %s not found", packageFunction[0]));
    }

    private class AWSContext implements Context {
        @Override
        public String getAwsRequestId() {
            return null;
        }

        @Override
        public String getLogGroupName() {
            return null;
        }

        @Override
        public String getLogStreamName() {
            return null;
        }

        @Override
        public String getFunctionName() {
            return "IRON_LAMBDA";
        }

        @Override
        public String getFunctionVersion() {
            return null;
        }

        @Override
        public String getInvokedFunctionArn() {
            return null;
        }

        @Override
        public CognitoIdentity getIdentity() {
            return null;
        }

        @Override
        public ClientContext getClientContext() {
            return null;
        }

        @Override
        public int getRemainingTimeInMillis() {
            return 0;
        }

        @Override
        public int getMemoryLimitInMB() {
            return 0;
        }

        @Override
        public LambdaLogger getLogger() {
            return null;
        }
    }
}


//---------payload helper-------//

class PayloadHelper {
    public static final Gson gson = new Gson();

    private PayloadHelper() {
    }

    public static boolean isJSONValid(String JSON_STRING) {
        if (!maybeJSON(JSON_STRING)) return false;
        try {
            gson.fromJson(JSON_STRING, Object.class);
            return true;
        } catch (com.google.gson.JsonSyntaxException ex) {
            return false;
        }
    }

    public static boolean maybeJSON(String test) {
        return (Objects.equals(String.valueOf(test.charAt(0)), "{") || Objects.equals(String.valueOf(test.charAt(0)), "[")) &&
                (Objects.equals(test.substring(test.length() - 1), "}") || Objects.equals(test.substring(test.length() - 1), "]"));
    }

    public static boolean isInteger(String s) {
        return isInteger(s, 10);
    }

    public static boolean isInteger(String s, int radix) {
        if (s.isEmpty()) return false;
        for (int i = 0; i < s.length(); i++) {
            if (i == 0 && s.charAt(i) == '-') {
                if (s.length() == 1) return false;
                else continue;
            }
            if (Character.digit(s.charAt(i), radix) < 0) return false;
        }
        return true;
    }

}