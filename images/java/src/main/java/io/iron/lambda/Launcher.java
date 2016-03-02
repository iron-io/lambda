package io.iron.lambda;

import java.io.*;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.*;

import io.github.lukehutch.fastclasspathscanner.FastClasspathScanner;

public class Launcher {
    public static void main(String[] args) {
        String handler = System.getenv("handler");
        String payload = System.getenv("payload");
        try {
            String[] packageMethod = validateInputParamsAndGetPackageMethod(handler, payload);
            Launcher ll = new Launcher();
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

    private void launchMethod(String[] packageHandler, String payload) throws Exception {
        boolean processed = false;
        boolean methodFound = false;
        Object result = null;
        String packageClass = packageHandler[0];
        String handlerName = packageHandler[1];

        AWSContext aws_ctx = new AWSContext();

        Class cls = Class.forName(packageClass); //get class in package
        Object lambdaClass = cls.newInstance();

        String packageName = lambdaClass.getClass().getPackage().getName();

        List<String> classNames = new FastClasspathScanner(packageName).scan().getNamesOfAllClasses();
        String packageNamePrefix = packageName + ".";

        Method[] declaredMethods = cls.getDeclaredMethods();

        for (Method lambdaMethod : declaredMethods) {
            if (Objects.equals(lambdaMethod.getName(), handlerName)) {  //if method name in user class == method name in env var
                System.out.println(String.format("Found package: %s and method: %s", packageClass, handlerName));
                methodFound = true;

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
                } else if (checkIfLambdaMethodRequiredPOJO(parameterTypes, classNames, packageNamePrefix) || checkIfLambdaMethodRequiredMap(parameterTypes) || checkIfLambdaMethodRequiredList(parameterTypes)) {
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
        if (!methodFound) {
            System.out.println(String.format("Method %s not found", handlerName));
            System.exit(1);
        }
        System.out.println(processed ? String.format("Method %s executed with result: %s", handlerName, (String) result) :
                String.format("Handler %s with simple, POJO, or IO(input/output) types not found", handlerName));
    }

    public boolean checkIfLambdaMethodRequiredIOStreams(Class[] parameterTypes) {
        return parameterTypes.length == 3 && parameterTypes[0].getTypeName().equals("java.io.InputStream") && parameterTypes[1].getTypeName().equals("java.io.OutputStream");
    }

    public boolean checkIfLambdaMethodRequiredPOJO(Class[] parameterTypes, List<String> classNames, String packageNamePrefix) {
        boolean classExist = false;
        for (String className : classNames) {
            if (className.startsWith(packageNamePrefix) && className.toLowerCase().contains(parameterTypes[0].getTypeName().toLowerCase())) {
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
}
