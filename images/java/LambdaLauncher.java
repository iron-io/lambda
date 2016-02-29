import java.io.*;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.*;

import com.amazonaws.services.lambda.runtime.ClientContext;
import com.amazonaws.services.lambda.runtime.CognitoIdentity;
import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.LambdaLogger;

import com.google.gson.Gson;

public class LambdaLauncher {
    public static void main(String[] args) {
        String handler = System.getenv("handler");
        String payload = System.getenv("payload");
        try {
            String[] packageMethod = validateInputParamsAndGetPackageMethod(handler, payload);
            LambdaLauncher ll = new LambdaLauncher();
            ll.launchMethod(packageMethod, payload);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static String[] validateInputParamsAndGetPackageMethod(String handler, String payload) {
        if (handler == null) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        if (payload == null || payload.equals("")) {
            System.out.println("Payload is empty, please specify payload");
            System.exit(1);
        }
        String[] package_function = handler.split("::");
        if (package_function[0] == null || package_function[0].equals("") || package_function[1] == null || package_function[1].equals("")) {
            System.out.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        return package_function;
    }

    public void launchMethod(String[] packageHandler, String payload) throws Exception {
        boolean processed = false;
        boolean classFound = false;
        Object result = null;
        String packageName = packageHandler[0];
        String handlerName = packageHandler[1];

        AWSContext aws_ctx = new AWSContext();

        Class cls = Class.forName(packageName); //get class in package
        Object lambdaClass = cls.newInstance();

        Method[] declaredMethods = cls.getDeclaredMethods();

        for (Method lambdaMethod : declaredMethods) {
            if (Objects.equals(lambdaMethod.getName(), handlerName)) {  //if method name in user class == method name in env var
                System.out.println(String.format("Found package: %s and method: %s", packageName, handlerName));
                classFound = true;

                Class[] parameterTypes = lambdaMethod.getParameterTypes();
                if (!parameterTypes[parameterTypes.length - 1].getTypeName().equals("com.amazonaws.services.lambda.runtime.Context")) {
                    System.out.println("The last param in method must be AWS context");
                    System.exit(1);
                }
                // IOStreams
                if (checkIfLambdaMethodRequiredIOStreams(parameterTypes)) {
                    ByteArrayOutputStream out = new ByteArrayOutputStream();
                    lambdaMethod.invoke(lambdaClass, new ByteArrayInputStream(payload.getBytes(StandardCharsets.UTF_8)), out, aws_ctx);
                    result = new String(out.toByteArray(), "UTF-8");
                    processed = true;
                // POJO, map, list
                } else if (checkIfLambdaMethodRequiredPOJO(parameterTypes, cls.getClasses()) || checkIfLambdaMethodRequiredMap(parameterTypes) || checkIfLambdaMethodRequiredList(parameterTypes)) {
                    result = ClassTypeHelper.gson.toJson(lambdaMethod.invoke(lambdaClass, ClassTypeHelper.gson.fromJson(payload, parameterTypes[0]), aws_ctx));
                    processed = true;
                // int, string, bool
                } else if (checkIfLambdaMethodRequiredPrimitiveType(parameterTypes)) {
                    if (Objects.equals(parameterTypes[0].toString(), "int") || Objects.equals(parameterTypes[0].toString(), "Integer")) {
                        result = lambdaMethod.invoke(lambdaClass, Integer.parseInt(payload), aws_ctx);
                        processed = true;
                    } else if (parameterTypes[0].getName().equals("java.lang.String")) {
                        result = lambdaMethod.invoke(lambdaClass, payload, aws_ctx);
                        processed = true;
                    } else if (Objects.equals(parameterTypes[0].toString(), "boolean") || Objects.equals(parameterTypes[0].toString(), "Boolean")) {
                        result = lambdaMethod.invoke(lambdaClass, Boolean.valueOf(payload), aws_ctx);
                        processed = true;
                    }
                }
            }
        }
        if (!classFound) {
            System.out.println(String.format("Class %s not found", packageName));
            System.exit(1);
        }
        System.out.println(processed ? String.format("Method %s executed with result: %s", handlerName, (String) result) :
                String.format("Handler %s with simple, POJO, or IO(input/output) types not found", handlerName));
    }

    public boolean checkIfLambdaMethodRequiredIOStreams(Class[] parameterTypes) {
        return parameterTypes.length == 3 && parameterTypes[0].getTypeName().equals("java.io.InputStream") && parameterTypes[1].getTypeName().equals("java.io.OutputStream");
    }

    public boolean checkIfLambdaMethodRequiredPOJO(Class[] parameterTypes, Class[] classes) {
        boolean classExist = false;
        for (Class clazz : classes) {
            if (Objects.equals(clazz.getName(), parameterTypes[0].getTypeName())) {
                classExist = true;
                break;
            }
        }
        return parameterTypes.length == 2 && classExist;
    }

    private boolean checkIfLambdaMethodRequiredMap(Class[] parameterTypes) {
        return parameterTypes.length == 2 && (parameterTypes[0].getTypeName().equals("java.util.Map") || parameterTypes[0].getTypeName().equals("java.util.HashMap"));
    }

    private boolean checkIfLambdaMethodRequiredList(Class[] parameterTypes) {
        return parameterTypes.length == 2 && (parameterTypes[0].getTypeName().equals("java.util.ArrayList") || parameterTypes[0].getTypeName().equals("java.util.List"));
    }
    private boolean checkIfLambdaMethodRequiredPrimitiveType(Class[] parameterTypes) {
        return parameterTypes.length == 2 && ClassTypeHelper.isSimpleType(parameterTypes[0]);
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
            return null;
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



class ClassTypeHelper {
    public static final Gson gson = new Gson();
    private static final Set<Class<?>> CLASS_TYPES = getSimpleTypes();

    private ClassTypeHelper() {}

    public static boolean isSimpleType(Class<?> classType) {
        return CLASS_TYPES.contains(classType);
    }

    private static Set<Class<?>> getSimpleTypes() {
        Set<Class<?>> ret = new HashSet<>();
        ret.add(String.class);
        ret.add(Integer.class);
        ret.add(int.class);
        ret.add(Boolean.class);
        return ret;
    }

}