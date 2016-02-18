import java.lang.reflect.Method;
import java.util.Objects;

import com.amazonaws.services.lambda.runtime.ClientContext;
import com.amazonaws.services.lambda.runtime.CognitoIdentity;
import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.LambdaLogger;

import com.google.gson.Gson;

public class LambdaLaunchder {
    public static void main(String[] args) {
        String handler_env = System.getenv("handler");
        String payload = System.getenv("payload");

        if (handler_env == null) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        if (payload == null || payload.equals("")) {
            System.out.println("Payload is empty");
            System.exit(1);
        }

        String[] package_function = handler_env.split("::");
        if (package_function[0] == null || package_function[0].equals("") || package_function[1] == null || package_function[1].equals("")) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        try {
            LambdaLaunchder ll = new LambdaLaunchder();
            ll.launch_function(package_function, payload);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    public void launch_function(String[] packageFunction, String payload) throws Exception {
//        Class[] paramInt = new Class[1];
//        paramInt[0] = Integer.TYPE;
//        Class[] params = new Class[2];
//        params[0] = Integer.TYPE;
//        params[1] = Context.class;
        String payloadType = "string";
        if (PayloadHelper.isInteger(payload)) {
            payloadType = "int";
            int payload_type = Integer.parseInt(payload);
            System.out.println("PAYLOAD IS INTEGER!");
        } else if (PayloadHelper.isJSONValid(payload)) {
            payloadType = "json";
            //TODO: implement POJO (gson already included)
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
                System.out.println("Found package: " + packageFunction[0] + " and method: " + packageFunction[1]);
                Class[] parameterTypes = method.getParameterTypes();
                Method lambdaMethod = cls.getDeclaredMethod(packageFunction[1], parameterTypes);

                for (Class parameterType : parameterTypes) {
                    //TODO: check first and second param only, second or third must be context
                    if (Objects.equals(payloadType, "int") && (Objects.equals(parameterType.toString(), payloadType) || Objects.equals(parameterType.toString(), "Integer"))) {
                        System.out.println(lambdaMethod.invoke(lambdaClass, Integer.parseInt(payload), aws_ctx));
                    } else if (Objects.equals(payloadType, "int") && (Objects.equals(parameterType.toString(), payloadType) || Objects.equals(parameterType.toString(), "Integer"))){
                        System.out.println(lambdaMethod.invoke(lambdaClass, payload, aws_ctx));
                    }
                }
            }

//            System.out.println(aM.getName());
//            Class[] parameterTypes = aM.getParameterTypes();
//            for (Class parameterType : parameterTypes) {
//                System.out.println(parameterType.toString());
//            }
        }

//        Method method = cls.getDeclaredMethod("test", paramInt);
//        method.invoke(obj, 123);
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
    private static final Gson gson = new Gson();

    private PayloadHelper() {
    }

    public static boolean isJSONValid(String JSON_STRING) {
        try {
            gson.fromJson(JSON_STRING, Object.class);
            return true;
        } catch (com.google.gson.JsonSyntaxException ex) {
            return false;
        }
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