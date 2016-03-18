package io.iron.lambda;

import java.util.UUID;

import com.amazonaws.services.lambda.runtime.ClientContext;
import com.amazonaws.services.lambda.runtime.CognitoIdentity;
import com.amazonaws.services.lambda.runtime.LambdaLogger;
import com.amazonaws.services.lambda.runtime.Context;

public class AWSContext implements Context {
    private String functionName;
    private String functionVersion;
    private String requestId;
    private long endTime;
    private int memory;

    // Constructs a context from environment variables.
    public AWSContext() {
        init();
    }

    private void init() {
        functionName = System.getenv("AWS_LAMBDA_FUNCTION_NAME");
        if (functionName == null) {
            functionName = "";
        }

        functionVersion = System.getenv("AWS_LAMBDA_FUNCTION_VERSION");
        if (functionVersion == null) {
            functionVersion = "$LATEST";
        }

        requestId = System.getenv("TASK_ID");
        if (requestId == null) {
            requestId = UUID.randomUUID().toString();
        }

        figureOutTime();
        figureOutMemory();
    }

    private void figureOutTime() {
        long startTime = System.currentTimeMillis();
        String timeoutEnv = System.getenv("TASK_TIMEOUT");
        if (timeoutEnv == null) {
            timeoutEnv = "";
        }
  
        long timeout = 3600;
        try {
            timeout = Integer.parseInt(timeoutEnv);
        } catch (Exception e) {
        }
  
        endTime = startTime + timeout*1000;
    }

    private void figureOutMemory() {
        String memEnv = System.getenv("TASK_MAXRAM");
        if (memEnv == null) {
            memEnv = "";
        }

        int bytes = 300 * 1024 * 1024;

        if (memEnv.length() > 0 && Character.isDigit(memEnv.charAt(memEnv.length() - 1))) {
            try {
              bytes = Integer.parseInt(memEnv);
            } catch(Exception e) {
            }
        } else if (memEnv.length() > 0) {
            char unit = memEnv.charAt(memEnv.length() - 1);
            try {
                int value = Integer.parseInt(memEnv.substring(0, memEnv.length() - 1));
                int multiplier = -1;
                switch (unit) {
                    case 'b': multiplier = 1; break;
                    case 'k': multiplier = 1024; break;
                    case 'm': multiplier = 1024 * 1024; break;
                    case 'g': multiplier = 1024 * 1024 * 1024; break;
                }

                if (multiplier >= 0) {
                    bytes = value * multiplier;
                }
            } catch(Exception e) {
            }

            memory = bytes / 1024 / 1024;
        }
    }

    @Override
    public String getAwsRequestId() {
        return requestId;
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
        return functionName;
    }

    @Override
    public String getFunctionVersion() {
        return functionVersion;
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
        int duration = (int)(endTime - System.currentTimeMillis());
        return Math.max(duration, 0);
    }

    @Override
    public int getMemoryLimitInMB() {
        return memory;
    }

    @Override
    public LambdaLogger getLogger() {
        return null;
    }
}
